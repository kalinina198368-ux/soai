<template>
  <div class="page-teaching">
    <div class="page-header">
      <h2 class="page-title">教学</h2>
      <p class="page-desc">
        按步骤生成口播脚本、语音、生图提示词与素材；配图每张生成完即显示；左侧可查看历史记录，任意环节均可重新生成并自动保存草稿。
      </p>
    </div>

    <div class="teaching-layout">
      <aside class="records-aside">
        <div class="aside-head">
          <span>教学记录</span>
          <el-button type="primary" link @click="newDraft">新建</el-button>
        </div>
        <el-scrollbar class="records-scroll">
          <div v-if="recordsLoading" class="aside-muted">加载中…</div>
          <div
            v-else
            v-for="r in recordsList"
            :key="r.id"
            class="record-row"
            :class="{ active: recordId === r.id }"
            @click="loadRecord(r.id)"
          >
            <div class="record-title">{{ r.title_summary || r.topic || "未命名" }}</div>
            <div class="record-meta">{{ formatRecordTime(r.updated_at) }}</div>
          </div>
          <div v-if="!recordsLoading && recordsList.length === 0" class="aside-muted">暂无记录，生成脚本后会写入草稿。</div>
        </el-scrollbar>
      </aside>

      <div class="teaching-main">
        <el-card class="steps-card" shadow="never">
          <el-steps :active="step" finish-status="success" align-center>
            <el-step title="口播脚本" description="LLM" />
            <el-step title="语音合成" description="TTS" />
            <el-step title="生图提示词" description="LLM" />
            <el-step title="素材生成" description="绘图 API" />
          </el-steps>
          <div v-if="recordId" class="view-mode-bar">
            <el-switch v-model="showAllStages" inline-prompt active-text="一览全部" inactive-text="分步浏览" />
            <span class="view-mode-hint">历史记录默认展开全部环节；关闭后仅显示当前步骤，可用底部上/下一步切换。</span>
          </div>
        </el-card>

        <!-- 1 口播脚本 -->
        <el-card v-show="showAllStages || step === 0" class="panel" shadow="never">
          <template #header>
            <div class="panel-header-row">
              <div class="panel-header-left">
                <span class="card-head">主题与口播脚本</span>
                <el-tag size="small" type="success">服务端</el-tag>
              </div>
              <el-button
                v-if="scriptPayload"
                type="primary"
                link
                :loading="loadingScript"
                :disabled="!topic.trim()"
                @click="regenScript"
              >
                重新生成脚本
              </el-button>
            </div>
          </template>
          <el-form label-position="top">
            <el-form-item label="教学主题">
              <el-input
                v-model="topic"
                placeholder="例如：线性回归、古诗背诵技巧、产品需求文档怎么写"
                maxlength="80"
                show-word-limit
                clearable
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="loadingScript" :disabled="!topic.trim()" @click="onGenerateScript">
                生成口播脚本文案
              </el-button>
              <span class="hint">调用 /api/teaching/script，使用后台配置的「教学文案模型」</span>
            </el-form-item>
          </el-form>

          <div v-if="scriptPayload" class="script-block">
            <div class="script-block-toolbar">
              <span class="script-toolbar-label">口播稿</span>
              <div class="script-toolbar-actions">
                <template v-if="!scriptEditMode">
                  <el-tooltip content="自定义编辑标题与各段正文" placement="top">
                    <el-button :icon="Edit" circle size="small" type="primary" plain @click="scriptEditMode = true" />
                  </el-tooltip>
                </template>
                <template v-else>
                  <el-button :icon="Check" size="small" type="success" @click="onFinishScriptEdit">完成编辑</el-button>
                </template>
              </div>
            </div>

            <template v-if="!scriptEditMode">
              <div class="script-title">{{ scriptPayload.title }}</div>
              <div v-for="(block, idx) in scriptPayload.outline" :key="block.id || 'seg-' + idx" class="script-seg">
                <div class="seg-title">{{ block.title }}</div>
                <div class="seg-text">{{ block.text }}</div>
              </div>
            </template>
            <template v-else>
              <el-form label-position="top" class="script-edit-form">
                <el-form-item label="标题">
                  <el-input v-model="scriptPayload.title" maxlength="120" show-word-limit />
                </el-form-item>
                <el-form-item
                  v-for="(block, idx) in scriptPayload.outline"
                  :key="block.id || 'edit-seg-' + idx"
                  :label="'第 ' + (idx + 1) + ' 段'"
                >
                  <el-input v-model="block.title" placeholder="小节标题" maxlength="80" class="seg-title-input" />
                  <el-input v-model="block.text" type="textarea" :rows="5" placeholder="口播正文" maxlength="8000" show-word-limit />
                </el-form-item>
              </el-form>
            </template>
          </div>
        </el-card>

        <!-- 2 语音 -->
        <el-card v-show="showAllStages || step === 1" class="panel" shadow="never">
          <template #header>
            <div class="panel-header-row">
              <div class="panel-header-left">
                <span class="card-head">AI 声音</span>
                <el-tag size="small" type="success">服务端</el-tag>
              </div>
              <el-button v-if="scriptPayload" type="primary" link :disabled="ttsLoading" @click="regenTts">重新合成语音</el-button>
            </div>
          </template>
          <p class="muted">调用 /api/teaching/tts，使用字节 OpenSpeech 配置；音频保存到本地存储目录。</p>
          <div v-if="ttsLoading" class="tts-loading">
            <el-progress :percentage="ttsProgress" :stroke-width="10" />
            <p class="muted">正在合成语音… {{ ttsVoiceName }}</p>
          </div>
          <div v-else-if="ttsItems.length" class="tts-list">
            <el-table :data="ttsItems" size="small" border stripe>
              <el-table-column prop="segmentId" label="分段" width="100" />
              <el-table-column prop="fileName" label="文件名" />
              <el-table-column prop="durationSec" label="时长(秒)" width="100" />
              <el-table-column label="试听" width="220">
                <template #default="{ row }">
                  <audio v-if="row.url" :src="teachingMediaSrc(row.url)" controls style="height: 32px; max-width: 200px" />
                  <span v-else class="muted">—</span>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-card>

        <!-- 3 生图提示词 -->
        <el-card v-show="showAllStages || step === 2" class="panel" shadow="never">
          <template #header>
            <div class="panel-header-row">
              <div class="panel-header-left">
                <span class="card-head">静态配图 · AI 生图提示词</span>
                <el-tag size="small" type="success">服务端</el-tag>
              </div>
              <el-button v-if="scriptPayload" type="primary" link @click="regenImagePrompts">重新生成提示词</el-button>
            </div>
          </template>
          <p class="muted">
            调用 /api/teaching/image-prompts：分镜密度与「一节一图」等以<strong>管理后台 → 系统配置 → 教学</strong>中的「配图颗粒度」为准；长稿会分批请求大模型。
          </p>
          <div v-for="(row, prIdx) in imagePromptRows" :key="row.segmentId || row.id || 'pr-' + prIdx" class="prompt-card">
            <div class="prompt-card-head">
              <span>{{ row.title }}</span>
              <el-tag size="small">{{ row.aspectRatio }}</el-tag>
            </div>
            <el-input type="textarea" :rows="4" readonly :model-value="row.prompt" />
          </div>
        </el-card>

        <!-- 4 素材 -->
        <el-card v-show="showAllStages || step === 3" class="panel" shadow="never">
          <template #header>
            <div class="panel-header-row">
              <div class="panel-header-left">
                <span class="card-head">教学视频素材</span>
                <el-tag size="small" type="success">配图 API</el-tag>
              </div>
            </div>
          </template>
          <p class="muted">
            每张单独请求 /api/teaching/images，生成一张即显示；子分镜（如 seg-1__2）与父段共用同一段 TTS 试听。失败可对该张重试。
          </p>
          <div class="asset-grid">
            <div v-for="task in assetTasks" :key="task.segmentId" class="asset-item">
              <div class="asset-item-top">
                <span>{{ task.label }}</span>
                <div class="asset-item-actions">
                  <el-tag v-if="task.status === 'done'" type="success" size="small">完成</el-tag>
                  <el-tag v-else-if="task.status === 'running'" type="warning" size="small">生成中</el-tag>
                  <el-tag v-else-if="task.status === 'error'" type="danger" size="small">失败</el-tag>
                  <el-tag v-else type="info" size="small">等待</el-tag>
                  <el-button type="primary" link size="small" :disabled="task.status === 'running'" @click="regenOneImage(task.segmentId)">
                    重新生成
                  </el-button>
                </div>
              </div>
              <el-progress :percentage="task.progress" :status="task.status === 'done' ? 'success' : undefined" />
              <p v-if="task.errorMsg" class="asset-error">{{ task.errorMsg }}</p>
              <div v-if="task.previewImage" class="asset-preview">
                <el-image
                  :src="teachingMediaSrc(task.previewImage)"
                  fit="cover"
                  class="thumb"
                  lazy
                  referrerpolicy="no-referrer"
                >
                  <template #error>
                    <div class="thumb-error">
                      <a :href="task.previewImage" target="_blank" rel="noopener noreferrer">无法预览，打开原图</a>
                    </div>
                  </template>
                </el-image>
                <div class="audio-line">
                  <audio v-if="task.audioUrl" :src="teachingMediaSrc(task.audioUrl)" controls style="height: 28px; max-width: 160px" />
                  <span>{{ task.audioLabel }}</span>
                </div>
              </div>
            </div>
          </div>
        </el-card>

        <div v-show="!showAllStages" class="footer-actions">
          <el-button :disabled="step === 0" @click="onPrev">上一步</el-button>
          <el-button type="primary" :disabled="!canNext" :loading="nextLoading" @click="onNext">
            {{ step === 3 ? "完成" : "下一步" }}
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { ElMessage } from "element-plus";
import { Check, Edit } from "@element-plus/icons-vue";
import { httpGet, httpPost } from "@/utils/http";

const step = ref(0);
const topic = ref("");
const loadingScript = ref(false);
const scriptPayload = ref(null);
const scriptEditMode = ref(false);

const ttsLoading = ref(false);
const ttsProgress = ref(0);
const ttsVoiceName = ref("ByteDance OpenSpeech");
const ttsItems = ref([]);

const imagePromptRows = ref([]);

const assetTasks = ref([]);

const recordId = ref(0);
const recordsList = ref([]);
const recordsLoading = ref(false);
/** 历史记录加载后默认 true，避免 max_step=3 时只能看到配图一步 */
const showAllStages = ref(false);

let saveTimer = null;

const unwrap = (res) => {
  if (res == null) return null;
  if (typeof res.code !== "undefined" && res.code !== 0) {
    const err = new Error(res.message || "请求失败");
    throw err;
  }
  return res.data != null ? res.data : res;
};

/** 仅保留有口播正文的段落（用户可删掉中间/末尾段或留空占位） */
const outlineWithBodyText = (outline) => (outline || []).filter((b) => (b.text || "").trim().length > 0);

const apiBaseTrim = () => (process.env.VUE_APP_API_HOST || "").replace(/\/$/, "");

/**
 * 教学配图/音频地址：以 / 开头的静态资源拼到 API 主机；HTTPS 页面加载 HTTP 外链时走 /api/download 代理（避免混合内容拦截）。
 */
const teachingMediaSrc = (url) => {
  if (!url || typeof url !== "string") return "";
  const u = url.trim();
  if (!u) return "";
  const api = apiBaseTrim();

  if (u.startsWith("/") && !u.startsWith("//")) {
    return api ? `${api}${u}` : u;
  }

  try {
    const abs = new URL(u);
    const needProxy =
      typeof window !== "undefined" &&
      window.location?.protocol === "https:" &&
      abs.protocol === "http:";
    if (needProxy) {
      const enc = encodeURIComponent(abs.href);
      if (api) {
        try {
          const base = new URL(api.startsWith("//") ? `${window.location.protocol}${api}` : api);
          if (base.protocol === "https:") {
            return `${api.replace(/\/$/, "")}/api/download?url=${enc}`;
          }
        } catch {
          /* 使用同源相对路径 */
        }
      }
      return `/api/download?url=${enc}`;
    }
  } catch {
    return u;
  }

  return u;
};

const scheduleSaveRecord = () => {
  if (saveTimer) clearTimeout(saveTimer);
  saveTimer = setTimeout(() => {
    saveTimer = null;
    saveRecord();
  }, 900);
};

const formatRecordTime = (unix) => {
  if (!unix) return "";
  return new Date(unix * 1000).toLocaleString();
};

const fetchRecords = async () => {
  recordsLoading.value = true;
  try {
    const res = await httpGet("/api/teaching/record/list");
    const d = unwrap(res);
    recordsList.value = d.list || [];
  } catch {
    recordsList.value = [];
  } finally {
    recordsLoading.value = false;
  }
};

const saveRecord = async () => {
  if (!topic.value.trim() && !recordId.value) return;
  if (!scriptPayload.value && !recordId.value) return;
  try {
    const assetsPayload = assetTasks.value
      .filter((t) => t.previewImage)
      .map((t) => ({ segmentId: t.segmentId, imageUrl: t.previewImage }));
    const res = await httpPost("/api/teaching/record/save", {
      id: recordId.value || 0,
      topic: topic.value.trim(),
      script: scriptPayload.value || {},
      tts_items: ttsItems.value || [],
      image_prompts: imagePromptRows.value || [],
      assets: assetsPayload,
      max_step: step.value,
    });
    const d = unwrap(res);
    if (d && d.id) recordId.value = d.id;
    fetchRecords();
  } catch {
    /* 静默失败，避免打断创作 */
  }
};

const normalizeJsonArrayLike = (raw, isSingleRow) => {
  if (raw == null) return [];
  if (Array.isArray(raw)) return raw.filter((x) => x && typeof x === "object");
  if (typeof raw === "object") {
    if (isSingleRow(raw)) return [raw];
    const vals = Object.values(raw).filter((x) => x && typeof x === "object");
    if (vals.length) return vals;
  }
  return [];
};

const parseLoadedScript = (raw) => {
  if (raw == null) return null;
  let s = raw;
  if (typeof raw === "string") {
    try {
      s = JSON.parse(raw);
    } catch {
      return null;
    }
  }
  if (typeof s !== "object" || s == null) return null;
  let outline = Array.isArray(s.outline) ? s.outline : [];
  outline = outline.map((b) => {
    if (!b || typeof b !== "object") return { id: "", title: "", text: "" };
    const id = b.id ?? b.segmentId ?? b.segment_id ?? "";
    return {
      id,
      segmentId: b.segmentId ?? b.segment_id ?? id,
      title: typeof b.title === "string" ? b.title : "",
      text: typeof b.text === "string" ? b.text : "",
    };
  });
  const title = typeof s.title === "string" ? s.title : "";
  if (!title && !outline.length) return null;
  return { title, outline };
};

/** 后端 assets 为 JSON 数组；若为 {}、单条对象或按 key 存放的对象，统一成数组 */
const normalizeTeachingAssets = (raw) => {
  if (raw == null) return [];
  if (Array.isArray(raw)) return raw.filter((a) => a && typeof a === "object");
  if (typeof raw === "object") {
    if (raw.segmentId != null && raw.imageUrl != null) return [raw];
    const vals = Object.values(raw).filter((x) => x && typeof x === "object" && x.segmentId != null);
    if (vals.length) return vals;
  }
  return [];
};

const segKey = (v) => (v == null || v === "" ? "" : String(v));

/** 子分镜 id 如 seg-1__2 对应 TTS 父段 seg-1 */
const parentSegmentIdForTts = (segmentId) => {
  const s = segKey(segmentId);
  if (!s) return "";
  const i = s.indexOf("__");
  if (i > 0) return s.slice(0, i);
  return s;
};

const ttsRowForShot = (segmentId) => {
  const k = segKey(segmentId);
  const by = Object.fromEntries(ttsItems.value.map((t) => [segKey(t.segmentId), t]));
  return by[k] || by[segKey(parentSegmentIdForTts(k))] || {};
};

const hydrateAssets = (assetsRaw) => {
  const assetsArr = normalizeTeachingAssets(assetsRaw);
  const prompts = imagePromptRows.value
    .filter((r) => (r.prompt || "").trim())
    .map((r) => ({
      segmentId: r.segmentId || r.id,
      prompt: r.prompt,
    }));
  const byAsset = Object.fromEntries(assetsArr.map((a) => [segKey(a.segmentId), a]));
  const rowMeta = Object.fromEntries(imagePromptRows.value.map((r) => [segKey(r.segmentId || r.id), r]));
  assetTasks.value = prompts.map((p) => {
    const k = segKey(p.segmentId);
    const a = byAsset[k];
    const tts = ttsRowForShot(p.segmentId);
    const meta = rowMeta[k] || {};
    const done = !!(a && a.imageUrl);
    return {
      segmentId: p.segmentId,
      label: meta.title || p.segmentId,
      previewImage: a?.imageUrl || "",
      audioLabel: tts?.fileName || "",
      audioUrl: tts?.url || "",
      progress: done ? 100 : 0,
      status: done ? "done" : "pending",
      errorMsg: "",
    };
  });
};

/** 接口或旧数据把 JSON 再包成字符串时，解成对象/数组 */
const unwrapJsonValue = (raw) => {
  let v = raw;
  for (let i = 0; i < 6; i++) {
    if (v == null) return null;
    if (typeof v !== "string") return v;
    const t = v.trim();
    if (!t) return null;
    if (t[0] !== "[" && t[0] !== "{") return v;
    try {
      v = JSON.parse(t);
    } catch {
      return raw;
    }
  }
  return v;
};

const loadRecord = async (id) => {
  try {
    const res = await httpGet(`/api/teaching/record/${id}`);
    const d = unwrap(res);
    recordId.value = d.id;
    topic.value = d.topic || "";
    scriptPayload.value = parseLoadedScript(unwrapJsonValue(d.script));
    ttsItems.value = normalizeJsonArrayLike(
      unwrapJsonValue(d.tts_items),
      (o) => (o.segmentId != null && o.segmentId !== "") || o.url != null || o.fileName != null
    );
    imagePromptRows.value = normalizeJsonArrayLike(
      unwrapJsonValue(d.image_prompts),
      (o) => (o.prompt != null && String(o.prompt).trim() !== "") || o.segmentId != null || o.id != null
    );
    hydrateAssets(unwrapJsonValue(d.assets));
    const ms = typeof d.max_step === "number" ? d.max_step : 0;
    step.value = Math.min(Math.max(ms, 0), 3);
    scriptEditMode.value = false;
    showAllStages.value = true;
    ElMessage.success("已加载记录");
  } catch (e) {
    ElMessage.error(e.message || String(e));
  }
};

const newDraft = () => {
  recordId.value = 0;
  topic.value = "";
  scriptPayload.value = null;
  imagePromptRows.value = [];
  ttsItems.value = [];
  assetTasks.value = [];
  step.value = 0;
  scriptEditMode.value = false;
  ttsProgress.value = 0;
  ttsLoading.value = false;
  showAllStages.value = false;
};

const onFinishScriptEdit = () => {
  scriptEditMode.value = false;
  ElMessage.success("已保存修改，后续步骤将使用当前文案");
  scheduleSaveRecord();
};

const onGenerateScript = async () => {
  if (!topic.value.trim()) {
    ElMessage.warning("请先填写主题");
    return;
  }
  loadingScript.value = true;
  try {
    const res = await httpPost("/api/teaching/script", { topic: topic.value.trim() });
    const d = unwrap(res);
    scriptPayload.value = { title: d.title, outline: d.outline || [] };
    scriptEditMode.value = false;
    imagePromptRows.value = [];
    ttsItems.value = [];
    assetTasks.value = [];
    ElMessage.success("口播脚本已生成");
    scheduleSaveRecord();
  } catch (e) {
    ElMessage.error(e.message || String(e));
  } finally {
    loadingScript.value = false;
  }
};

const regenScript = async () => {
  if (!topic.value.trim()) {
    ElMessage.warning("请先填写主题");
    return;
  }
  scriptPayload.value = null;
  imagePromptRows.value = [];
  ttsItems.value = [];
  assetTasks.value = [];
  await onGenerateScript();
};

const runFakeTts = async () => {
  if (!scriptPayload.value?.outline?.length) return;
  const segments = outlineWithBodyText(scriptPayload.value.outline).map((o) => ({
    id: o.id || o.segmentId,
    text: o.text || "",
  }));
  if (!segments.length) {
    ElMessage.warning("请至少保留一节有口播正文的段落后再合成语音");
    return;
  }
  ttsLoading.value = true;
  ttsProgress.value = 5;
  ttsItems.value = [];
  try {
    const res = await httpPost("/api/teaching/tts", { segments });
    const d = unwrap(res);
    ttsItems.value = (d.items || []).map((it) => ({
      segmentId: it.segmentId,
      fileName: it.fileName,
      url: it.url,
      durationSec: it.durationSec ?? 0,
    }));
    ttsProgress.value = 100;
    ElMessage.success("语音合成完成");
    scheduleSaveRecord();
  } catch (e) {
    ElMessage.error(e.message || String(e));
  } finally {
    ttsLoading.value = false;
  }
};

const regenTts = async () => {
  await runFakeTts();
};

const loadImagePrompts = async () => {
  if (!scriptPayload.value?.outline?.length) return;
  const bodyOutline = outlineWithBodyText(scriptPayload.value.outline);
  if (!bodyOutline.length) {
    ElMessage.warning("请至少保留一节有口播正文的段落后再生成配图提示词");
    return;
  }
  try {
    const res = await httpPost("/api/teaching/image-prompts", {
      topic: topic.value.trim(),
      outline: bodyOutline,
    });
    const d = unwrap(res);
    imagePromptRows.value = d.prompts || [];
    if (!imagePromptRows.value.length) {
      ElMessage.warning("未返回生图提示词");
    } else {
      scheduleSaveRecord();
    }
  } catch (e) {
    ElMessage.error(e.message || String(e));
  }
};

const regenImagePrompts = async () => {
  imagePromptRows.value = [];
  assetTasks.value = [];
  await loadImagePrompts();
};

const patchTaskBySegment = (segmentId, partial) => {
  const idx = assetTasks.value.findIndex((t) => t.segmentId === segmentId);
  if (idx < 0) return;
  const cur = assetTasks.value[idx];
  assetTasks.value.splice(idx, 1, { ...cur, ...partial });
};

const runParallelAssets = async () => {
  const prompts = imagePromptRows.value
    .filter((r) => (r.prompt || "").trim())
    .map((r) => ({
      segmentId: r.segmentId || r.id,
      prompt: r.prompt,
    }));
  if (!prompts.length) {
    ElMessage.warning("没有可生成的配图提示词");
    return;
  }
  const rowMeta = Object.fromEntries(imagePromptRows.value.map((r) => [segKey(r.segmentId || r.id), r]));
  assetTasks.value = prompts.map((p) => {
    const k = segKey(p.segmentId);
    const tts = ttsRowForShot(p.segmentId);
    return {
      segmentId: p.segmentId,
      label: rowMeta[k]?.title || p.segmentId,
      previewImage: "",
      audioLabel: tts?.fileName || "",
      audioUrl: tts?.url || "",
      progress: 5,
      status: "running",
      errorMsg: "",
    };
  });

  await Promise.all(
    prompts.map(async (p) => {
      try {
        const res = await httpPost("/api/teaching/images", { prompts: [p] });
        const d = unwrap(res);
        const list = normalizeTeachingAssets(d.assets);
        const a = list[0];
        if (!a || !a.imageUrl) throw new Error("未返回配图地址");
        const sk = segKey(p.segmentId);
        const meta = rowMeta[sk] || {};
        const tts = ttsRowForShot(p.segmentId);
        patchTaskBySegment(p.segmentId, {
          label: meta.title || p.segmentId,
          previewImage: a.imageUrl,
          audioLabel: tts?.fileName || "",
          audioUrl: tts?.url || "",
          progress: 100,
          status: "done",
          errorMsg: "",
        });
        scheduleSaveRecord();
      } catch (e) {
        patchTaskBySegment(p.segmentId, {
          status: "error",
          errorMsg: e.message || String(e),
          progress: 0,
          previewImage: "",
        });
      }
    })
  );

  const allOk = assetTasks.value.length && assetTasks.value.every((t) => t.status === "done");
  const anyOk = assetTasks.value.some((t) => t.status === "done");
  if (allOk) ElMessage.success("配图已全部生成");
  else if (anyOk) ElMessage.warning("部分配图已生成，失败项可点击「重新生成」重试");
  else ElMessage.error("配图生成失败，请检查配图接口配置后重试");
  scheduleSaveRecord();
};

const regenOneImage = async (segmentId) => {
  const row = imagePromptRows.value.find((r) => (r.segmentId || r.id) === segmentId);
  if (!row || !(row.prompt || "").trim()) {
    ElMessage.warning("未找到对应提示词");
    return;
  }
  const p = { segmentId: row.segmentId || row.id, prompt: row.prompt };
  const meta = row;
  const tts = ttsRowForShot(p.segmentId);
  patchTaskBySegment(segmentId, {
    status: "running",
    progress: 40,
    errorMsg: "",
    previewImage: "",
    label: meta.title || p.segmentId,
    audioLabel: tts?.fileName || "",
    audioUrl: tts?.url || "",
  });
  try {
    const res = await httpPost("/api/teaching/images", { prompts: [p] });
    const d = unwrap(res);
    const list = normalizeTeachingAssets(d.assets);
    const a = list[0];
    if (!a || !a.imageUrl) throw new Error("未返回配图地址");
    patchTaskBySegment(segmentId, {
      previewImage: a.imageUrl,
      progress: 100,
      status: "done",
      errorMsg: "",
    });
    ElMessage.success("该张配图已更新");
    scheduleSaveRecord();
  } catch (e) {
    patchTaskBySegment(segmentId, {
      status: "error",
      errorMsg: e.message || String(e),
      progress: 0,
    });
    ElMessage.error(e.message || String(e));
  }
};

const nextLoading = ref(false);

const canNext = computed(() => {
  if (step.value === 0) return !!scriptPayload.value;
  if (step.value === 1) return !ttsLoading.value && ttsItems.value.length > 0;
  if (step.value === 2) return imagePromptRows.value.length > 0;
  if (step.value === 3) {
    return assetTasks.value.length > 0 && assetTasks.value.every((t) => t.status !== "running");
  }
  return false;
});

watch(step, (s) => {
  if (s !== 0) {
    scriptEditMode.value = false;
  }
  if (s === 1 && scriptPayload.value && !ttsLoading.value && ttsItems.value.length === 0) {
    runFakeTts();
  }
  if (s === 2 && imagePromptRows.value.length === 0) {
    loadImagePrompts();
  }
});

const onPrev = () => {
  if (step.value <= 0) return;
  step.value -= 1;
};

const onNext = async () => {
  if (step.value === 0 && !scriptPayload.value) {
    ElMessage.info("请先生成口播脚本");
    return;
  }
  if (step.value === 1 && (ttsLoading.value || !ttsItems.value.length)) {
    ElMessage.info("请等待语音合成完成");
    return;
  }
  if (step.value === 2 && !imagePromptRows.value.length) {
    await loadImagePrompts();
    if (!imagePromptRows.value.length) return;
  }
  if (step.value === 3) {
    if (assetTasks.value.some((t) => t.status === "running")) {
      ElMessage.info("仍有配图在生成中，请稍候");
      return;
    }
    if (!assetTasks.value.length) {
      ElMessage.info("暂无配图任务，请确认已生成提示词");
      return;
    }
    ElMessage.success("本流程已结束，左侧可继续查看或新建记录");
    scheduleSaveRecord();
    return;
  }
  if (step.value === 2) {
    nextLoading.value = true;
    step.value = 3;
    await runParallelAssets();
    nextLoading.value = false;
    return;
  }
  step.value += 1;
};

onMounted(() => {
  fetchRecords();
});
</script>
<style scoped lang="stylus">
.page-teaching
  max-width 1200px
  margin 0 auto
  padding 16px 20px 100px
  color var(--text-color, #303133)

.teaching-layout
  display flex
  align-items flex-start
  gap 20px

.records-aside
  flex 0 0 220px
  position sticky
  top 72px
  max-height calc(100vh - 100px)
  border-radius 12px
  border 1px solid var(--border-color, #ebeef5)
  background var(--card-bg, #fff)
  overflow hidden

.aside-head
  display flex
  align-items center
  justify-content space-between
  padding 10px 12px
  font-size 13px
  font-weight 600
  border-bottom 1px solid var(--border-color, #ebeef5)

.records-scroll
  height calc(100vh - 200px)
  max-height 520px

.aside-muted
  font-size 12px
  opacity 0.65
  padding 10px 12px

.record-row
  padding 10px 12px
  cursor pointer
  border-bottom 1px solid var(--border-color, rgba(0,0,0,0.04))
  transition background 0.15s

.record-row:hover
  background var(--chat-bg, rgba(0,0,0,0.03))

.record-row.active
  background rgba(117, 79, 246, 0.08)

.record-title
  font-size 13px
  font-weight 500
  line-height 1.4
  overflow hidden
  text-overflow ellipsis
  display -webkit-box
  -webkit-line-clamp 2
  -webkit-box-orient vertical

.record-meta
  font-size 11px
  opacity 0.6
  margin-top 4px

.teaching-main
  flex 1
  min-width 0

.page-header
  margin-bottom 16px

.page-title
  margin 0 0 8px
  font-size 22px
  font-weight 600
  color var(--text-theme-color, #754ff6)

.page-desc
  margin 0
  font-size 13px
  opacity 0.85
  line-height 1.6

.steps-card
  margin-bottom 16px
  background var(--card-bg, #fff)
  border-radius 12px

.view-mode-bar
  margin-top 14px
  padding-top 12px
  border-top 1px solid var(--border-color, #ebeef5)
  display flex
  flex-direction column
  align-items flex-start
  gap 8px

.view-mode-hint
  font-size 12px
  line-height 1.5
  opacity 0.72
  max-width 720px
  margin-bottom 16px
  background var(--card-bg, #fff)
  border-radius 12px

.panel-header-row
  display flex
  align-items center
  justify-content space-between
  gap 12px
  flex-wrap wrap

.panel-header-left
  display flex
  align-items center
  gap 8px

.card-head
  font-weight 600

.hint
  margin-left 12px
  font-size 12px
  opacity 0.7

.script-block
  margin-top 12px
  padding 12px
  border-radius 8px
  background var(--chat-bg, rgba(0,0,0,0.03))

.script-block-toolbar
  display flex
  align-items center
  justify-content space-between
  margin-bottom 12px
  padding-bottom 8px
  border-bottom 1px solid var(--border-color, rgba(0,0,0,0.06))

.script-toolbar-label
  font-size 13px
  font-weight 600
  color var(--text-theme-color, #754ff6)

.script-toolbar-actions
  display flex
  align-items center
  gap 8px

.script-edit-form
  :deep(.el-form-item__label)
    font-weight 500

.script-edit-form .seg-title-input
  margin-bottom 8px

.script-title
  font-weight 600
  margin-bottom 12px

.script-seg
  margin-bottom 12px

.seg-title
  font-size 13px
  font-weight 600
  margin-bottom 6px
  color var(--text-theme-color, #754ff6)

.seg-text
  font-size 14px
  line-height 1.65
  white-space pre-wrap

.muted
  font-size 13px
  opacity 0.75
  margin 0 0 12px

.tts-loading
  max-width 520px

.tts-list
  margin-top 8px

.prompt-card
  margin-bottom 14px
  padding 12px
  border-radius 8px
  border 1px solid var(--border-color, #ebeef5)

.prompt-card-head
  display flex
  align-items center
  justify-content space-between
  margin-bottom 8px
  font-weight 500

.asset-grid
  display flex
  flex-direction column
  gap 14px

.asset-item
  padding 12px
  border-radius 8px
  border 1px solid var(--border-color, #ebeef5)
  background var(--chat-bg, rgba(0,0,0,0.02))

.asset-item-top
  display flex
  align-items center
  justify-content space-between
  margin-bottom 8px
  font-size 13px
  gap 8px

.asset-item-actions
  display flex
  align-items center
  gap 8px
  flex-shrink 0

.asset-error
  font-size 12px
  color var(--el-color-danger)
  margin 6px 0 0

.asset-preview
  display flex
  align-items center
  gap 12px
  margin-top 10px

.thumb
  width 120px
  height 68px
  border-radius 6px

.thumb-error
  width 120px
  min-height 48px
  padding 6px
  font-size 11px
  line-height 1.4
  display flex
  align-items center
  justify-content center
  text-align center
  background var(--el-fill-color-light)
  border-radius 6px

  a
    color var(--el-color-primary)
    word-break break-all

.audio-line
  font-size 12px
  opacity 0.8

.footer-actions
  position sticky
  bottom 0
  display flex
  justify-content flex-end
  gap 12px
  padding 12px 0
  background linear-gradient(to top, var(--chat-bg, #f5f7fa) 60%, transparent)

@media (max-width: 900px)
  .teaching-layout
    flex-direction column
  .records-aside
    position relative
    top 0
    max-height none
    width 100%
    flex none
  .records-scroll
    max-height 220px
</style>
