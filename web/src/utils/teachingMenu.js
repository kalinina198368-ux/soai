/** 教学功能：前端兜底菜单项（若后台已配置同路径则不再追加） */
export const TEACHING_MENU_ENTRY = {
  name: "教学",
  icon: "icon-book",
  url: "/teaching",
  sort_num: 100,
};

export function withTeachingMenu(items) {
  const list = Array.isArray(items) ? [...items] : [];
  const norm = (u) => String(u || "").replace(/\/$/, "") || "";
  if (list.some((x) => norm(x.url || x.URL) === "/teaching")) {
    return list;
  }
  return [...list, { ...TEACHING_MENU_ENTRY }];
}
