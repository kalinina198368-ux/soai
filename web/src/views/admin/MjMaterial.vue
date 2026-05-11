<template>
  <div class="container mj-material">
    <el-tabs v-model="activeTab" @tab-change="onTabChange" class="material-tabs">
      <el-tab-pane label="图片素材" name="image"></el-tab-pane>
      <el-tab-pane label="视频素材" name="video"></el-tab-pane>
    </el-tabs>

    <el-card class="category-card" v-loading="categoryLoading">
      <div class="card-header">
        <span>素材分类</span>
        <el-button size="small" type="primary" :icon="Plus" @click="openCategoryDialog()">新增分类</el-button>
      </div>
      <el-table
        class="category-table"
        :data="categories"
        :row-key="(row) => row.id"
        table-layout="auto"
      >
        <el-table-column prop="title" label="分类名称">
          <template #default="{ row }">
            <span class="sort" :data-id="row.id">
              <i class="iconfont icon-drag"></i>
              {{ row.title }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="内部标识" />
        <el-table-column prop="is_active" label="启用">
          <template #default="{ row }">
            <el-switch v-model="row.is_active" @change="toggleCategory(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="170">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="openCategoryDialog(row)">编辑</el-button>
            <el-popconfirm title="确定要删除当前分类吗？" :width="210" @confirm="removeCategory(row)">
              <template #reference>
                <el-button size="small" type="danger">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card class="material-card" v-loading="materialLoading">
      <div class="card-header">
        <span>素材列表</span>
        <div class="header-actions">
          <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
            <el-form-item label="分类">
              <el-select v-model="filters.category_id" clearable placeholder="全部分类" filterable>
                <el-option v-for="cat in categories" :key="cat.id" :label="cat.title" :value="cat.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="名称">
              <el-input v-model="filters.title" placeholder="标题/名称" clearable />
            </el-form-item>
            <el-form-item label="类型" v-if="activeTab === 'image'">
              <el-select v-model="filters.type" clearable placeholder="全部">
                <el-option v-for="opt in typeOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
              </el-select>
            </el-form-item>
            <el-form-item label="状态">
              <el-select v-model="filters.is_active" clearable placeholder="全部">
                <el-option label="已启用" :value="true" />
                <el-option label="已停用" :value="false" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :icon="Search" @click="fetchMaterials()">查询</el-button>
            </el-form-item>
            <el-form-item>
              <el-button @click="resetFilters">重置</el-button>
            </el-form-item>
          </el-form>
          <el-button type="primary" :icon="Plus" @click="openMaterialDialog()">新增素材</el-button>
        </div>
      </div>

      <el-table
        class="material-table"
        :data="materials"
        :row-key="(row) => row.id"
        table-layout="auto"
      >
        <el-table-column prop="title" label="标题">
          <template #default="{ row }">
            <span class="sort" :data-id="row.id">
              <i class="iconfont icon-drag"></i>
              {{ row.title }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="标识" />
        <el-table-column prop="category_id" label="分类">
          <template #default="{ row }">
            {{ categoryName(row.category_id) }}
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="110" v-if="activeTab === 'image'">
          <template #default="{ row }">
            <el-tag size="small" type="info">
              {{ typeLabel(row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="image" label="素材" width="120" v-if="activeTab === 'image'">
          <template #default="{ row }">
            <el-image
              v-if="row.image"
              :src="row.image"
              style="width: 80px; height: 80px; object-fit: cover;"
              fit="cover"
              :preview-src-list="[row.image]"
            />
            <el-tag v-else type="info">无</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="video" label="视频" width="120" v-if="activeTab === 'video'">
          <template #default="{ row }">
            <video
              v-if="row.video"
              :src="row.video"
              style="width: 80px; height: 80px; object-fit: cover;"
              controls
              muted
            />
            <el-tag v-else type="info">无</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="image" label="封面图" width="120" v-if="activeTab === 'video'">
          <template #default="{ row }">
            <el-image
              v-if="row.image"
              :src="row.image"
              style="width: 80px; height: 80px; object-fit: cover;"
              fit="cover"
              :preview-src-list="[row.image]"
            />
            <el-tag v-else type="info">无</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="prompt" label="提示词" min-width="180">
          <template #default="{ row }">
            <el-tooltip v-if="row.prompt" class="box-item" effect="dark" :content="row.prompt" placement="top-start">
              <span class="prompt">{{ row.prompt }}</span>
            </el-tooltip>
            <el-tag v-else type="info">无</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="启用" width="90">
          <template #default="{ row }">
            <el-switch v-model="row.is_active" @change="toggleMaterial(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="openMaterialDialog(row)">编辑</el-button>
            <el-popconfirm title="确定要删除该素材吗？" :width="200" @confirm="removeMaterial(row)">
              <template #reference>
                <el-button size="small" type="danger">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination" v-if="materials.length">
        <el-pagination
          background
          layout="total, prev, pager, next, sizes"
          :total="page.total || 0"
          :current-page="page.page"
          :page-size="page.page_size"
          :page-sizes="[10, 20, 50]"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>

    <!-- 分类弹窗 -->
    <el-dialog
      v-model="showCategoryDialog"
      :title="categoryDialogTitle"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form :model="categoryForm" label-width="100px" ref="categoryFormRef" :rules="categoryRules">
        <el-form-item label="分类名称" prop="title">
          <el-input v-model="categoryForm.title" placeholder="展示名称" />
        </el-form-item>
        <el-form-item label="内部标识" prop="name">
          <el-input v-model="categoryForm.name" placeholder="英文/唯一标识" />
        </el-form-item>
        <el-form-item label="排序值">
          <el-input-number v-model="categoryForm.sort_order" :min="1" :step="1" />
        </el-form-item>
        <el-form-item label="启用状态">
          <el-switch v-model="categoryForm.is_active" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showCategoryDialog = false">取消</el-button>
          <el-button type="primary" @click="saveCategory">提交</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 素材弹窗 -->
    <el-dialog
      v-model="showMaterialDialog"
      :title="materialDialogTitle"
      width="700px"
      :close-on-click-modal="false"
    >
      <el-form :model="materialForm" label-width="110px" ref="materialFormRef" :rules="getMaterialRules()">
        <el-form-item label="所属分类" prop="category_id">
          <el-select v-model="materialForm.category_id" filterable placeholder="选择分类">
            <el-option v-for="cat in categories" :key="cat.id" :label="cat.title" :value="cat.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="标题" prop="title">
          <el-input v-model="materialForm.title" />
        </el-form-item>
        <el-form-item label="标识" prop="name">
          <el-input v-model="materialForm.name" />
        </el-form-item>
        <el-form-item label="提示词">
          <el-input v-model="materialForm.prompt" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="类型" prop="type" v-if="activeTab === 'image'">
          <el-select v-model="materialForm.type" placeholder="请选择">
            <el-option v-for="opt in typeOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
          </el-select>
        </el-form-item>
        <el-form-item :label="activeTab === 'image' ? '素材图' : '视频'" v-if="activeTab === 'image'">
          <el-input v-model="materialForm.image" placeholder="素材地址">
            <template #append>
              <el-upload :show-file-list="false" :auto-upload="true" :http-request="(file)=>uploadImg(file, 'image')">
                上传
              </el-upload>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="视频" v-if="activeTab === 'video'">
          <el-input v-model="materialForm.video" placeholder="视频地址">
            <template #append>
              <el-upload :show-file-list="false" :auto-upload="true" :http-request="(file)=>uploadVideo(file, 'video')">
                上传
              </el-upload>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="封面图" v-if="activeTab === 'video'">
          <el-input v-model="materialForm.image" placeholder="封面图地址">
            <template #append>
              <el-upload :show-file-list="false" :auto-upload="true" :http-request="(file)=>uploadImg(file, 'image')">
                上传
              </el-upload>
            </template>
          </el-input>
        </el-form-item>
        <!-- <el-form-item label="预览图">
          <el-input v-model="materialForm.preview" placeholder="预览图地址">
            <template #append>
              <el-upload :show-file-list="false" :auto-upload="true" :http-request="(file)=>uploadImg(file, 'preview')">
                上传
              </el-upload>
            </template>
          </el-input>
        </el-form-item> -->
        <el-form-item label="排序值">
          <el-input-number v-model="materialForm.sort_order" :min="1" :step="1" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="materialForm.is_active" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showMaterialDialog = false">取消</el-button>
          <el-button type="primary" @click="saveMaterial">提交</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { nextTick, onMounted, onUnmounted, reactive, ref } from "vue";
import { httpGet, httpPost } from "@/utils/http";
import { ElMessage } from "element-plus";
import { removeArrayItem } from "@/utils/libs";
import { Plus, Search } from "@element-plus/icons-vue";
import Compressor from "compressorjs";
import { Sortable } from "sortablejs";

const typeOptions = [
  { label: "文生图", value: "txt2img" },
  { label: "图生图", value: "img2img" },
  { label: "通用", value: "all" },
];

// 当前标签页：image 图片素材，video 视频素材
const activeTab = ref("image");

// 获取 API 基础路径
const getApiBase = () => {
  return activeTab.value === "image" ? "/api/admin/mj/material" : "/api/admin/sora/material";
};

// 分类数据
const categories = ref([]);
const categoryLoading = ref(true);
const showCategoryDialog = ref(false);
const categoryDialogTitle = ref("新增分类");
const categoryForm = ref({ is_active: true, sort_order: 1 });
const categoryFormRef = ref(null);
const categoryRules = reactive({
  title: [{ required: true, message: "请输入分类名称", trigger: "blur" }],
  name: [{ required: true, message: "请输入内部标识", trigger: "blur" }],
});
let categorySortable;

// 素材数据
const materials = ref([]);
const page = reactive({ page: 1, page_size: 10, total: 0 });
const filters = reactive({
  category_id: "",
  title: "",
  type: "",
  is_active: "",
});
const materialLoading = ref(true);
const showMaterialDialog = ref(false);
const materialDialogTitle = ref("新增素材");
const materialForm = ref({
  category_id: "",
  title: "",
  name: "",
  prompt: "",
  type: "all",
  image: "",
  video: "",
  is_active: true,
  sort_order: 1,
});
const materialFormRef = ref(null);
const materialRules = reactive({
  category_id: [{ required: true, message: "请选择分类", trigger: "change" }],
  title: [{ required: true, message: "请输入标题", trigger: "blur" }],
  name: [{ required: true, message: "请输入标识", trigger: "blur" }],
  type: [{ required: true, message: "请选择类型", trigger: "change" }],
});

// 动态获取验证规则
const getMaterialRules = () => {
  const rules = {
    category_id: [{ required: true, message: "请选择分类", trigger: "change" }],
    title: [{ required: true, message: "请输入标题", trigger: "blur" }],
    name: [{ required: true, message: "请输入标识", trigger: "blur" }],
  };
  if (activeTab.value === "image") {
    rules.type = [{ required: true, message: "请选择类型", trigger: "change" }];
  }
  return rules;
};
let materialSortable;

const normalizeCategory = (list) =>
  (list || []).map((item, index) => ({
    ...item,
    is_active: item.enable ?? item.is_active ?? false,
    sort_order: item.sort ?? item.sort_order ?? index + 1,
  }));

const normalizeMaterial = (list) =>
  (list || []).map((item, index) => ({
    ...item,
    is_active: item.is_active ?? item.enable ?? false,
    sort_order: item.sort_order ?? index + 1,
  }));

const fetchCategories = () => {
  categoryLoading.value = true;
  httpPost(`${getApiBase()}/category/list`, {})
    .then((res) => {
      categories.value = normalizeCategory(res.data);
      nextTick(bindCategorySortable);
    })
    .catch((e) => {
      ElMessage.error("获取分类失败：" + e.message);
    })
    .finally(() => {
      categoryLoading.value = false;
    });
};

const fetchMaterials = () => {
  materialLoading.value = true;
  const payload = {
    category_id: filters.category_id || 0,
    title: filters.title,
    type: filters.type,
    is_active: filters.is_active === "" ? undefined : filters.is_active,
    page: page.page,
    page_size: page.page_size,
  };
  if (payload.is_active === undefined) {
    delete payload.is_active;
  }
  // 视频素材不需要 type 字段
  if (activeTab.value === "video" && payload.type !== undefined) {
    delete payload.type;
  }
  httpPost(`${getApiBase()}/list`, payload)
    .then((res) => {
      const data = res.data || {};
      materials.value = normalizeMaterial(data.items);
      page.total = data.total || 0;
      page.page = data.page || payload.page;
      page.page_size = data.page_size || payload.page_size;
      nextTick(bindMaterialSortable);
    })
    .catch((e) => {
      ElMessage.error("获取素材失败：" + e.message);
    })
    .finally(() => {
      materialLoading.value = false;
    });
};

onMounted(() => {
  fetchCategories();
  fetchMaterials();
});

onUnmounted(() => {
  if (categorySortable) {
    categorySortable.destroy();
  }
  if (materialSortable) {
    materialSortable.destroy();
  }
});

const bindCategorySortable = () => {
  if (categorySortable) {
    categorySortable.destroy();
  }
  const body = document.querySelector(".category-table .el-table__body tbody");
  if (!body) return;
  categorySortable = Sortable.create(body, {
    animation: 200,
    onEnd({ newIndex, oldIndex, from }) {
      if (newIndex === oldIndex) return;
      const sortedIds = Array.from(from.children).map((row) =>
        parseInt(row.querySelector(".sort")?.getAttribute("data-id") || "0", 10)
      );
      const ids = [];
      const sorts = [];
      sortedIds.forEach((id, idx) => {
        ids.push(id);
        sorts.push(idx + 1);
        if (categories.value[idx]) {
          categories.value[idx].sort_order = idx + 1;
        }
      });
      httpPost(`${getApiBase()}/category/sort`, { ids, sorts }).catch((e) => {
        ElMessage.error("分类排序失败：" + e.message);
      });
    },
  });
};

const bindMaterialSortable = () => {
  if (materialSortable) {
    materialSortable.destroy();
  }
  const body = document.querySelector(".material-table .el-table__body tbody");
  if (!body) return;
  materialSortable = Sortable.create(body, {
    animation: 200,
    onEnd({ newIndex, oldIndex, from }) {
      if (newIndex === oldIndex) return;
      const sortedIds = Array.from(from.children).map((row) =>
        parseInt(row.querySelector(".sort")?.getAttribute("data-id") || "0", 10)
      );
      const ids = [];
      const sorts = [];
      sortedIds.forEach((id, idx) => {
        ids.push(id);
        sorts.push(idx + 1);
        if (materials.value[idx]) {
          materials.value[idx].sort_order = idx + 1;
        }
      });
      httpPost(`${getApiBase()}/sort`, { ids, sorts }).catch((e) => {
        ElMessage.error("素材排序失败：" + e.message);
      });
    },
  });
};

const openCategoryDialog = (row = null) => {
  categoryDialogTitle.value = row ? "编辑分类" : "新增分类";
  showCategoryDialog.value = true;
  categoryForm.value = row
    ? { ...row }
    : { title: "", name: "", is_active: true, sort_order: categories.value.length + 1 };
};

const saveCategory = () => {
  categoryFormRef.value.validate((valid) => {
    if (!valid) return;
    const payload = { ...categoryForm.value };
    if (!payload.sort_order) {
      payload.sort_order = categories.value.length + 1;
    }
    httpPost(`${getApiBase()}/category/save`, payload)
      .then(() => {
        ElMessage.success("分类保存成功");
        showCategoryDialog.value = false;
        fetchCategories();
      })
      .catch((e) => {
        ElMessage.error("分类保存失败：" + e.message);
      });
  });
};

const toggleCategory = (row) => {
  httpPost(`${getApiBase()}/category/enable`, { id: row.id, is_active: row.is_active }).catch((e) => {
    ElMessage.error("更新状态失败：" + e.message);
  });
};

const removeCategory = (row) => {
  httpGet(`${getApiBase()}/category/remove`, { id: row.id })
    .then(() => {
      ElMessage.success("分类已删除");
      categories.value = removeArrayItem(categories.value, row, (a, b) => a.id === b.id);
      if (filters.category_id === row.id) {
        filters.category_id = "";
        fetchMaterials();
      }
    })
    .catch((e) => {
      ElMessage.error("删除失败：" + e.message);
    });
};

const openMaterialDialog = (row = null) => {
  materialDialogTitle.value = row ? "编辑素材" : "新增素材";
  showMaterialDialog.value = true;
  materialForm.value = row
    ? { ...row }
    : {
        category_id: "",
        title: "",
        name: "",
        prompt: "",
        type: activeTab.value === "image" ? "all" : "",
        image: "",
        video: "",
        is_active: true,
        sort_order: materials.value.length + 1,
      };
};

const saveMaterial = () => {
  materialFormRef.value.validate((valid) => {
    if (!valid) return;
    const payload = { ...materialForm.value };
    payload.category_id = Number(payload.category_id);
    if (payload.created_at) {
      payload.created_at = Number(payload.created_at);
    }
    if (!payload.sort_order) {
      payload.sort_order = materials.value.length + 1;
    }
    // 视频素材不需要 type 字段
    if (activeTab.value === "video" && payload.type !== undefined) {
      delete payload.type;
    }
    httpPost(`${getApiBase()}/save`, payload)
      .then(() => {
        ElMessage.success("素材保存成功");
        showMaterialDialog.value = false;
        fetchMaterials();
      })
      .catch((e) => {
        ElMessage.error("素材保存失败：" + e.message);
      });
  });
};

const toggleMaterial = (row) => {
  httpPost(`${getApiBase()}/enable`, { id: row.id, is_active: row.is_active }).catch((e) => {
    ElMessage.error("更新状态失败：" + e.message);
  });
};

const removeMaterial = (row) => {
  httpGet(`${getApiBase()}/remove`, { id: row.id })
    .then(() => {
      ElMessage.success("素材已删除");
      materials.value = removeArrayItem(materials.value, row, (a, b) => a.id === b.id);
      page.total = Math.max(0, (page.total || 1) - 1);
    })
    .catch((e) => {
      ElMessage.error("删除失败：" + e.message);
    });
};

const uploadImg = (file, field) => {
  new Compressor(file.file, {
    quality: 0.6,
    success(result) {
      const formData = new FormData();
      formData.append("file", result, result.name);
      httpPost("/api/admin/upload", formData)
        .then((res) => {
          materialForm.value[field] = res.data.url;
          ElMessage.success("上传成功");
        })
        .catch((e) => {
          ElMessage.error("上传失败：" + e.message);
        });
    },
    error(err) {
      ElMessage.error("上传失败：" + err.message);
    },
  });
};

const categoryName = (id) => {
  const match = categories.value.find((c) => c.id === id);
  return match ? match.title : "-";
};

const typeLabel = (value) => {
  const opt = typeOptions.find((o) => o.value === value);
  return opt ? opt.label : value || "-";
};

const handleSizeChange = (size) => {
  page.page_size = size;
  page.page = 1;
  fetchMaterials();
};

const handlePageChange = (p) => {
  page.page = p;
  fetchMaterials();
};

const resetFilters = () => {
  filters.category_id = "";
  filters.title = "";
  filters.type = "";
  filters.is_active = "";
  page.page = 1;
  fetchMaterials();
};

// 标签页切换
const onTabChange = () => {
  // 重置筛选条件
  resetFilters();
  // 重新加载数据
  fetchCategories();
  fetchMaterials();
};

// 上传视频
const uploadVideo = (file, field) => {
  const formData = new FormData();
  formData.append("file", file.file, file.file.name);
  httpPost("/api/admin/upload", formData)
    .then((res) => {
      materialForm.value[field] = res.data.url;
      ElMessage.success("上传成功");
    })
    .catch((e) => {
      ElMessage.error("上传失败：" + e.message);
    });
};
</script>

<style lang="stylus" scoped>
.mj-material {
  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 15px;
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .filter-form {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .sort {
    cursor: move;
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }

  .prompt {
    max-width: 240px;
    display: inline-block;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    vertical-align: middle;
  }

  .pagination {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
  }

  .material-tabs {
    margin-bottom: 20px;
  }
}
</style>