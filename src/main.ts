import { Service } from "./mygo";
import "./style.css";

type Task = Awaited<ReturnType<typeof Service.list>>[number];
const app = document.querySelector<HTMLDivElement>("#app")!;
let selected = "";
let dialogOpen = false;
let lastClipboard = "";
let clipboardEnabled = localStorage.getItem("watchClipboard") !== "off";
let pollBusy = false;
let filenameRequest = 0;
let suggestedFilename = "";

app.innerHTML = [
 '<div class="shell">',
 '<header><div class="brand"><div class="mark">↓</div><div><h1>gdirect IDM</h1><p>Google Drive 公共链接下载器</p></div></div><div id="status" class="state">初始化中</div></header>',
 '<nav><button id="add" class="primary">＋ 新建下载</button><button id="resume">▶ 继续</button><button id="stop">■ 停止</button><button id="reveal">▣ 打开文件夹</button><label class="toggle"><input id="watch" type="checkbox"> 监听剪贴板</label></nav>',
 '<main><div class="section"><strong>下载任务</strong><span id="count">0 个任务</span></div><div class="tablewrap"><table><thead><tr><th>文件名</th><th>状态</th><th>已写入磁盘</th><th>保存路径</th></tr></thead><tbody id="jobs"></tbody></table></div></main>',
 '<footer>仅支持拥有公开下载权限的 Google Drive 单个文件。无需 OAuth。直连是否可用取决于当前网络。</footer>',
 '</div>',
 '<div id="overlay" class="overlay hidden"><section class="dialog" role="dialog" aria-modal="true" aria-labelledby="dialogTitle">',
 '<h2 id="dialogTitle">新建下载</h2><label for="url">Google Drive 文件分享链接</label><textarea id="url" rows="3" placeholder="https://drive.google.com/file/d/..."></textarea>',
 '<label for="path">保存地址及文件名</label><div class="saveRow"><input id="path" placeholder="点击浏览，选择目录并重命名"><button id="browse">浏览…</button></div>',
 '<p id="hint">请确认下载来源及保存位置</p><div class="actions"><button id="cancel">取消</button><button id="download" class="primary">开始下载</button></div>',
 '</section></div>'
].join("");

const el = (selector:string) => document.querySelector<HTMLElement>(selector)!;
const overlay = el("#overlay");
const urlInput = el("#url") as HTMLTextAreaElement;
const pathInput = el("#path") as HTMLInputElement;
const hint = el("#hint");
const watch = el("#watch") as HTMLInputElement;
watch.checked = clipboardEnabled;
function errorText(reason:unknown) {
 const text = reason instanceof Error ? reason.message : String(reason);
 hint.classList.add("error"); hint.textContent = text;
}
function showDialog(url = "") {
 dialogOpen = true;
 urlInput.value = url; pathInput.value = "";
 suggestedFilename = ""; pathInput.placeholder = "点击浏览，选择目录并重命名";
 hint.classList.remove("error");
 hint.textContent = "选择保存位置和文件名称，下载前不会自动执行";
 overlay.classList.remove("hidden"); urlInput.focus();
 void inspectFilename(url);
}
function hideDialog() { dialogOpen = false; filenameRequest++; overlay.classList.add("hidden"); }
async function inspectFilename(url: string) {
 const request = ++filenameRequest;
 suggestedFilename = "";
 if (!url.trim()) return;
 hint.classList.remove("error");
 hint.textContent = "正在读取原文件名（首次使用时可能需要扫描 IP）…";
 try {
  const name = await Service.suggestName(url.trim());
  if (request !== filenameRequest || !dialogOpen || urlInput.value.trim() !== url.trim()) return;
  suggestedFilename = name;
  hint.textContent = "已识别原文件名：" + name + "（可在保存窗口修改）";
  if (!pathInput.value) pathInput.placeholder = "默认名称：" + name;
 } catch (e) {
  if (request !== filenameRequest || !dialogOpen) return;
  hint.textContent = "无法自动读取原文件名，请手动命名并保留扩展名。";
  console.warn("Filename metadata unavailable",e);
 }
}
urlInput.addEventListener("change", () => { void inspectFilename(urlInput.value); });

el("#add").onclick = () => showDialog();
el("#cancel").onclick = hideDialog;
el("#browse").onclick = async () => {
 try { const path=await Service.choosePath(suggestedFilename || "GoogleDrive-file"); if(path)pathInput.value=path; } catch(e) {errorText(e);}
};
el("#download").onclick = async () => {
 try {
  hint.classList.remove("error");
  const url = urlInput.value.trim();
  await Service.validate(url);
  let path = pathInput.value.trim();
  if(!path){path = await Service.choosePath(suggestedFilename || "GoogleDrive-file");if(!path)return;pathInput.value=path;}
  const task = await Service.start(url,path);
  selected=task.id; hideDialog(); await refresh();
 } catch(e) {errorText(e);}
};
watch.onchange=()=>{clipboardEnabled=watch.checked;localStorage.setItem("watchClipboard", clipboardEnabled?"on":"off");};
async function refresh(){
 try {
  const tasks=await Service.list();
  el("#count").textContent=tasks.length+" 个任务";
  const tbody=el("#jobs");tbody.replaceChildren();
  if(tasks.length===0){const row=document.createElement("tr");const cell=document.createElement("td");cell.colSpan=4;cell.className="empty";cell.textContent="复制 Google Drive 公开文件链接，即可开始下载";row.append(cell);tbody.append(row);}
  for(const t of tasks as Task[]) {
   const row=document.createElement("tr");
   if(t.id===selected)row.className="selected";
   row.onclick=()=>{selected=t.id;void refresh();};
   const stateMap:Record<string,string>={running:"下载中",complete:"已完成",failed:"失败",stopped:"已停止"};
   const cells=[t.name,(stateMap[t.state]||t.state)+(t.error?" · "+t.error:""),t.bytes?(t.bytes/1048576).toFixed(1)+" MB":"—",t.path];
   cells.forEach((value,index)=>{const cell=document.createElement("td");cell.textContent=value;cell.title=value;if(index===3)cell.className="muted";row.append(cell);});
   tbody.append(row);
  }
 } catch(e){el("#status").textContent="后端连接失败";console.error(e);}
}
async function taskAction(action:"stop"|"resume"|"reveal") {
 if(!selected)return;
 try {
  const task=(await Service.list()).find(t=>t.id===selected);
  if(!task)return;
  if(action==="stop")await Service.stop(selected);
  if(action==="resume")await Service.resume(selected);
  if(action==="reveal")await Service.showInFolder(task.path);
  await refresh();
 } catch(e){alert(e instanceof Error?e.message:String(e));}
}
el("#stop").onclick=()=>{void taskAction("stop");};
el("#resume").onclick=()=>{void taskAction("resume");};
el("#reveal").onclick=()=>{void taskAction("reveal");};
async function poll(){
 if(pollBusy||!clipboardEnabled||dialogOpen)return;
 pollBusy=true;
 try{
  const value=(await Service.clipboardText()).trim();
  if(value!==lastClipboard) {
   lastClipboard=value;
   if(value){try{await Service.validate(value);showDialog(value);}catch{}}
  }
 }catch{}finally{pollBusy=false;}
}
(async()=>{
 try{el("#status").textContent=await Service.status();}catch{el("#status").textContent="后端不可用";}
 try{lastClipboard=(await Service.clipboardText()).trim();}catch{}
 await refresh();
 setInterval(()=>{void poll();},1200);
 setInterval(()=>{void refresh();},1200);
})();
