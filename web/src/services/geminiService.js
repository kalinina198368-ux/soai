// Gemini API 服务（参考 nb-app）
// 用于测试 B 供应商的 API

/**
 * 格式化 Gemini API 错误
 */
const formatGeminiError = (error) => {
  let message = "发生了未知错误，请稍后重试。";
  const errorMsg = error?.message || error?.toString() || "";

  if (errorMsg.includes("401") || errorMsg.includes("API key not valid")) {
    message = "API Key 无效或过期，请检查您的设置。";
  } else if (errorMsg.includes("403")) {
    message = "访问被拒绝。请检查您的网络连接（可能需要切换节点）或 API Key 权限。";
  } else if (errorMsg.includes("400")) {
    message = "请求参数无效 (400 Bad Request)。请检查您的设置或提示词。";
  } else if (errorMsg.includes("429")) {
    message = "请求过于频繁，请稍后再试（429 Too Many Requests）。";
  } else if (errorMsg.includes("503")) {
    message = "Gemini 服务暂时不可用，请稍后重试（503 Service Unavailable）。";
  } else if (errorMsg.includes("TypeError") || errorMsg.includes("Failed to fetch") || errorMsg.includes("NetworkError")) {
    message = "网络请求失败。可能是网络连接问题，或者请求内容过多（如图片太大）。";
  } else if (errorMsg.includes("SAFETY")) {
    message = "生成的内容因安全策略被拦截。请尝试修改您的提示词。";
  } else if (errorMsg.includes("404")) {
    message = "请求的模型不存在或路径错误 (404 Not Found)。";
  } else if (errorMsg.includes("500")) {
    message = "Gemini 服务器内部错误，请稍后重试 (500 Internal Server Error)。";
  } else {
    message = `请求出错: ${errorMsg}`;
  }

  const newError = new Error(message);
  newError.originalError = error;
  return newError;
};

/**
 * 构建用户内容
 */
const constructUserContent = (prompt, images) => {
  const userParts = [];
  
  // 添加图片
  images.forEach((img) => {
    userParts.push({
      inlineData: {
        mimeType: img.mimeType || 'image/png',
        data: img.base64Data,
      },
    });
  });

  // 添加文本提示词
  if (prompt && prompt.trim()) {
    userParts.push({ text: prompt });
  }

  return {
    role: "user",
    parts: userParts,
  };
};

/**
 * 将图片 URL 转换为 base64
 */
const imageUrlToBase64 = async (imageUrl) => {
  try {
    // 如果是 data URL，直接解析
    if (imageUrl.startsWith('data:')) {
      const [header, base64] = imageUrl.split(',');
      const mimeMatch = header.match(/data:([^;]+)/);
      const mimeType = mimeMatch ? mimeMatch[1] : 'image/png';
      return { base64Data: base64, mimeType };
    }

    // 如果是普通 URL，需要 fetch
    const response = await fetch(imageUrl);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const blob = await response.blob();
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onloadend = () => {
        const base64 = reader.result.split(',')[1]; // 移除 data:image/...;base64, 前缀
        const mimeType = blob.type || 'image/png';
        resolve({ base64Data: base64, mimeType });
      };
      reader.onerror = reject;
      reader.readAsDataURL(blob);
    });
  } catch (error) {
    throw new Error(`图片转换失败: ${error.message}`);
  }
};

/**
 * 生成图片内容（非流式）
 */
export const generateContent = async (
  apiKey,
  prompt,
  imageUrls = [],
  settings = {}
) => {
  try {
    // 动态导入 @google/genai
    const { GoogleGenAI } = await import("@google/genai");
    
    const ai = new GoogleGenAI({
      apiKey,
      httpOptions: {
        baseUrl: settings.customEndpoint || 'https://api.kuai.host'
      }
    });

    // 转换图片 URL 为 base64
    const images = [];
    for (const url of imageUrls) {
      try {
        const imageData = await imageUrlToBase64(url);
        images.push(imageData);
      } catch (error) {
        console.warn(`图片 ${url} 转换失败:`, error);
      }
    }

    // 构建用户内容
    const currentUserContent = constructUserContent(prompt, images);
    const contentsPayload = [currentUserContent];

    // 调用 API
    const response = await ai.models.generateContent({
      model: settings.modelName || "gemini-3-pro-image-preview",
      contents: contentsPayload,
      config: {
        imageConfig: {
          imageSize: settings.resolution || '2K',
          ...(settings.aspectRatio && settings.aspectRatio !== 'Auto' 
            ? { aspectRatio: settings.aspectRatio } 
            : {}),
        },
        tools: settings.useGrounding ? [{ googleSearch: {} }] : [],
        responseModalities: ["TEXT", "IMAGE"],
        ...(settings.enableThinking ? {
          thinkingConfig: {
            includeThoughts: true,
          }
        } : {}),
      },
    });

    const candidate = response.candidates?.[0];
    if (!candidate || !candidate.content || !candidate.content.parts) {
      throw new Error("No content generated.");
    }

    // 提取图片部分
    const imageParts = candidate.content.parts.filter(
      part => part.inlineData && !part.thought
    );

    // 将 base64 图片转换为可用的 URL
    const imageUrls_result = imageParts.map(part => {
      if (part.inlineData) {
        return `data:${part.inlineData.mimeType};base64,${part.inlineData.data}`;
      }
      return null;
    }).filter(url => url !== null);

    return {
      images: imageUrls_result,
      text: candidate.content.parts
        .filter(part => part.text && !part.thought)
        .map(part => part.text)
        .join('')
    };

  } catch (error) {
    console.error("Gemini API Error:", error);
    throw formatGeminiError(error);
  }
};

/**
 * 流式生成图片内容
 */
export const streamGeminiResponse = async function* (
  apiKey,
  prompt,
  imageUrls = [],
  settings = {}
) {
  try {
    // 动态导入 @google/genai
    const { GoogleGenAI } = await import("@google/genai");
    
    const ai = new GoogleGenAI({
      apiKey,
      httpOptions: {
        baseUrl: settings.customEndpoint || 'https://api.kuai.host'
      }
    });

    // 转换图片 URL 为 base64
    const images = [];
    for (const url of imageUrls) {
      try {
        const imageData = await imageUrlToBase64(url);
        images.push(imageData);
      } catch (error) {
        console.warn(`图片 ${url} 转换失败:`, error);
      }
    }

    // 构建用户内容
    const currentUserContent = constructUserContent(prompt, images);
    const contentsPayload = [currentUserContent];

    // 调用流式 API
    const responseStream = await ai.models.generateContentStream({
      model: settings.modelName || "gemini-3-pro-image-preview",
      contents: contentsPayload,
      config: {
        imageConfig: {
          imageSize: settings.resolution || '2K',
          ...(settings.aspectRatio && settings.aspectRatio !== 'Auto' 
            ? { aspectRatio: settings.aspectRatio } 
            : {}),
        },
        tools: settings.useGrounding ? [{ googleSearch: {} }] : [],
        responseModalities: ["TEXT", "IMAGE"],
        ...(settings.enableThinking ? {
          thinkingConfig: {
            includeThoughts: true,
          }
        } : {}),
      },
    });

    let currentParts = [];

    for await (const chunk of responseStream) {
      const candidates = chunk.candidates;
      if (!candidates || candidates.length === 0) continue;
      
      const newParts = candidates[0].content?.parts || [];

      for (const part of newParts) {
        const isThought = !!(part.thought);

        // 处理文本
        if (part.text !== undefined) {
          const lastPart = currentParts[currentParts.length - 1];
          if (
            lastPart && 
            lastPart.text !== undefined && 
            !!lastPart.thought === isThought
          ) {
            lastPart.text += part.text;
          } else {
            currentParts.push({ 
              text: part.text, 
              thought: isThought 
            });
          }
        } 
        // 处理图片
        else if (part.inlineData) {
          currentParts.push({ 
            inlineData: {
              mimeType: part.inlineData.mimeType || 'image/png',
              data: part.inlineData.data || ''
            }, 
            thought: isThought 
          });
        }
      }

      // 提取图片
      const imageParts = currentParts.filter(
        part => part.inlineData && !part.thought
      );
      const imageUrls_result = imageParts.map(part => {
        if (part.inlineData) {
          return `data:${part.inlineData.mimeType};base64,${part.inlineData.data}`;
        }
        return null;
      }).filter(url => url !== null);

      yield {
        images: imageUrls_result,
        text: currentParts
          .filter(part => part.text && !part.thought)
          .map(part => part.text)
          .join('')
      };
    }
  } catch (error) {
    console.error("Gemini API Stream Error:", error);
    throw formatGeminiError(error);
  }
};

