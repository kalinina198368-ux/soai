<template>
  <div class="container withdraw" v-loading="loading">
    <div class="handle-box">
      <el-input v-model="query.username" placeholder="用户名" class="handle-input mr10"></el-input>
      <el-select v-model="query.status" placeholder="提现状态" style="width: 120px">
        <el-option v-for="item in withdrawStatus" :key="item.value" :label="item.label" :value="item.value" />
      </el-select>
      <el-date-picker
        v-model="query.created_at"
        type="daterange"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
        format="YYYY-MM-DD"
        value-format="YYYY-MM-DD"
        style="margin: 0 10px; width: 200px; position: relative; top: 3px"
      />
      <el-button type="primary" :icon="Search" @click="fetchData">搜索</el-button>
    </div>

    <el-row>
      <el-table :data="items" :row-key="(row) => row.id" table-layout="auto">
        <el-table-column prop="id" label="申请ID" width="80" />
        <el-table-column prop="username" label="申请用户" width="120" />
        <el-table-column label="提现算力" width="100">
          <template #default="scope">
            <span style="color: #e74c3c; font-weight: 600;">{{ scope.row.power  }}</span>
          </template>
        </el-table-column>

       <el-table-column label="提现金额" width="100">
          <template #default="scope">
            <span style="color: #409eff; font-weight: 600;">{{scope.row.amount }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="account_name" label="收款人姓名" width="120" />
        <el-table-column label="收款码" width="100">
          <template #default="scope">
            <el-image
              v-if="scope.row.qrcode_url"
              :src="scope.row.qrcode_url"
              :preview-src-list="[scope.row.qrcode_url]"
              fit="cover"
              style="width: 50px; height: 50px; border-radius: 4px; cursor: pointer;"
            />
            <span v-else style="color: #999;">无</span>
          </template>
        </el-table-column>
        <el-table-column label="申请时间" width="160">
          <template #default="scope">
            <span>{{ dateFormat(scope.row["created_at"]) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="scope">
            <el-tag :type="getStatusType(scope.row.status)">
              {{ getStatusText(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" show-overflow-tooltip />

        <el-table-column label="操作" width="280" fixed="right">
          <template #default="scope">
            <el-button
              v-if="scope.row.status === 0"
              size="small"
              type="success"
              @click="handleApprove(scope.row)"
            >
              审核通过
            </el-button>
            <el-button
              v-if="scope.row.status === 0"
              size="small"
              type="warning"
              @click="handleReject(scope.row)"
            >
              拒绝
            </el-button>
            <el-button
              v-if="scope.row.status === 1"
              size="small"
              type="primary"
              @click="handlePaid(scope.row)"
            >
              已打款
            </el-button>
            <el-popconfirm title="确定要删除当前记录吗?" @confirm="remove(scope.row)">
              <template #reference>
                <el-button size="small" type="danger">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-row>

    <div class="pagination">
      <el-pagination
        v-if="total > 0"
        background
        layout="total,prev, pager, next"
        :hide-on-single-page="true"
        v-model:current-page="page"
        v-model:page-size="pageSize"
        @current-change="fetchData()"
        :total="total"
      />
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue";
import { httpGet, httpPost } from "@/utils/http";
import { ElMessage, ElMessageBox } from "element-plus";
import { dateFormat, removeArrayItem } from "@/utils/libs";
import { Delete, Search } from "@element-plus/icons-vue";

// 变量定义
const items = ref([]);
const query = ref({ username: "", created_at: [], status: -1 });
const total = ref(0);
const page = ref(1);
const pageSize = ref(15);
const loading = ref(true);
const withdrawStatus = ref([
  { value: -1, label: "全部" },
  { value: 0, label: "待审核" },
  { value: 1, label: "审核通过" },
  { value: 2, label: "已打款" },
  { value: 3, label: "已拒绝" },
]);

const getStatusText = (status) => {
  const statusMap = {
    0: "待审核",
    1: "审核通过",
    2: "已打款",
    3: "已拒绝",
  };
  return statusMap[status] || "未知";
};

const getStatusType = (status) => {
  const typeMap = {
    0: "warning",
    1: "info",
    2: "success",
    3: "danger",
  };
  return typeMap[status] || "";
};

onMounted(() => {
  fetchData();
});

// 获取数据
const fetchData = () => {
  loading.value = true;
  const params = {
    page: page.value,
    page_size: pageSize.value,
    username: query.value.username,
    status: query.value.status,
  };
  
  if (query.value.created_at && query.value.created_at.length === 2) {
    params.start_date = query.value.created_at[0];
    params.end_date = query.value.created_at[1];
  }
  
  httpPost("/api/admin/withdraw/list", params)
    .then((res) => {
      if (res.data) {
        items.value = res.data.items || res.data.data || [];
        total.value = res.data.total || 0;
        page.value = res.data.page || page.value;
        pageSize.value = res.data.page_size || pageSize.value;
      }
      loading.value = false;
    })
    .catch((e) => {
      loading.value = false;
      ElMessage.error("获取数据失败：" + e.message);
    });
};

const handleApprove = (row) => {
  ElMessageBox.confirm("确定要审核通过该提现申请吗?", "审核提示", {
    confirmButtonText: "确认",
    cancelButtonText: "取消",
    type: "warning",
  })
    .then(() => {
      httpPost("/api/admin/withdraw/approve", { id: row.id })
        .then(() => {
          ElMessage.success("审核通过成功");
          fetchData();
        })
        .catch((e) => {
          ElMessage.error("操作失败：" + e.message);
        });
    });
};

const handleReject = (row) => {
  ElMessageBox.prompt("请输入拒绝原因", "拒绝提现", {
    confirmButtonText: "确认",
    cancelButtonText: "取消",
    inputType: "textarea",
    inputPlaceholder: "请输入拒绝原因",
  })
    .then(({ value }) => {
      httpPost("/api/admin/withdraw/reject", { id: row.id, reason: value })
        .then(() => {
          ElMessage.success("已拒绝该提现申请");
          fetchData();
        })
        .catch((e) => {
          ElMessage.error("操作失败：" + e.message);
        });
    });
};

const handlePaid = (row) => {
  ElMessageBox.confirm("确认该提现已打款完成吗?", "打款确认", {
    confirmButtonText: "确认",
    cancelButtonText: "取消",
    type: "warning",
  })
    .then(() => {
      httpPost("/api/admin/withdraw/paid", { id: row.id })
        .then(() => {
          ElMessage.success("已标记为已打款");
          fetchData();
        })
        .catch((e) => {
          ElMessage.error("操作失败：" + e.message);
        });
    });
};

const remove = function (row) {
  httpGet("/api/admin/withdraw/remove?id=" + row.id)
    .then(() => {
      ElMessage.success("删除成功！");
      items.value = removeArrayItem(items.value, row, (v1, v2) => {
        return v1.id === v2.id;
      });
    })
    .catch((e) => {
      ElMessage.error("删除失败：" + e.message);
    });
};
</script>

<style lang="stylus" scoped>
.withdraw {
  .handle-box {
    margin-bottom 20px
    .handle-input {
      max-width 150px;
      margin-right 10px;
    }
  }

  .opt-box {
    padding-bottom: 10px;
    display flex;
    justify-content flex-end

    .el-icon {
      margin-right: 5px;
    }
  }

  .el-select {
    width: 100%
  }

  .pagination {
    margin-top: 20px;
    display flex
    justify-content right
  }
}
</style>
