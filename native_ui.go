package main

import (
 "fmt"
 "math"
 "os"
 "path/filepath"
 "strings"

 "github.com/egoist/mygo"
 "github.com/egoist/mygo/ui"
)

type downloadUI struct{
 service *Service
 win *mygo.Window
 tray *mygo.Tray
 items []Task
 list ui.ListState
 selected int
 showNew bool
 showDelete bool
 showRestart bool
 restartID string
 showRoute bool
 routeBusy bool
 routeChoice string
 routeCandidates []RouteOption
 deleteWithFiles bool
 deleteID string
 lastClipboard string
 url string
 filename string
 folder string
 inspection FileInspection
 inspectionSeq int
 inspecting bool
 notice string
}

func newDownloadUI(service *Service)*downloadUI{
 base,err:=os.UserHomeDir()
 if err!=nil{base="."}
 preferred:=filepath.Join(base,"Downloads")
 if st,err:=os.Stat(preferred);err!=nil||!st.IsDir(){preferred=base}
 return &downloadUI{service:service,selected:-1,routeChoice:"自动（推荐）",folder:preferred}
}

func (s *downloadUI) prompt(raw string){
 s.showNew=true
 s.url=strings.TrimSpace(raw)
 s.filename=""
 s.inspection=FileInspection{}
 s.inspecting=false
 s.notice=""
 s.inspectionSeq++
 if _,err:=s.service.Validate(s.url);err==nil{s.inspect()}
}

func (s *downloadUI) inspect(){
 raw:=strings.TrimSpace(s.url)
 if _,err:=s.service.Validate(raw);err!=nil{return}
 s.inspectionSeq++
 seq:=s.inspectionSeq
 s.inspecting=true
 go func(){
  info,err:=s.service.Inspect(raw)
  s.win.Update(func(){
   if !s.showNew||seq!=s.inspectionSeq||s.url!=raw{return}
   s.inspecting=false
   if err!=nil{
    s.notice="无法自动获取文件名，请填写文件名（含扩展名）："+err.Error()
    return
   }
   s.inspection=info
   if s.filename==""{s.filename=info.Name}
   s.notice="已读取原始文件名和公开文件信息"
  })
 }()
}


func (s *downloadUI) loadRoutes(refresh bool){
 s.win.Update(func(){s.routeBusy=true})
 go func(){
  var options []RouteOption
  var err error
  if refresh{options,err=s.service.RefreshRoutes()}else{options,err=s.service.ListRoutes()}
  s.win.Update(func(){
   s.routeBusy=false
   if err!=nil{s.notice="IP 列表获取失败："+err.Error();return}
   s.routeCandidates=options
   if s.routeChoice!="自动（推荐）"{
    exists:=false
    for _,option:=range options{if routeLabel(option)==s.routeChoice{exists=true;break}}
    if !exists{s.routeChoice="自动（推荐）"}
   }
   s.notice=fmt.Sprintf("已载入 %d 个可选 Google IP",len(options))
  })
 }()
}
func (s *downloadUI) chosenRoute()RouteOption{
 for _,option:=range s.routeCandidates{
  if routeLabel(option)==s.routeChoice{return option}
 }
 return RouteOption{}
}
func (s *downloadUI) routeModal(c *ui.Context){
 ui.Modal(c,&s.showRoute,func(){
  ui.Column(c).Width(520).Gap(10).Children(func(){
   ui.Text(c,"Google 下载线路").Bold().FontSize(18)
   ui.Text(c,"自动模式优先使用已验证的低延迟 IP，连接失败还可以切换备用 IP。").TextColor(c.Theme().TextMuted)
   options:=[]string{"自动（推荐）"}
   for _,r:=range s.routeCandidates{options=append(options,routeLabel(r))}
   ui.Text(c,"IP 选择")
   ui.Select(c,&s.routeChoice,options).Label("选择 Google IP")
   if s.routeBusy{ui.Progress(c,-1).Label("扫描 Google IP")}
   if s.routeChoice!="自动（推荐）"{
    ui.Text(c,"注意：手动固定 IP 不会自动故障切换，测速延迟不代表实际下载速度。").TextColor(c.Theme().TextMuted)
   }
   ui.Row(c).Gap(8).Children(func(){
    styleTactileButton(ui.Button(c,"重新扫描 IP").Disabled(s.routeBusy).OnClick(func(){s.loadRoutes(true)}),buttonSecondary,buttonMedium)
    ui.Spacer(c)
    styleTactileButton(ui.Button(c,"应用到选中任务").Disabled(s.selected<0||s.routeBusy).OnClick(func(){
     if t,ok:=s.selectedTask();ok{
      if t.State=="complete"{s.notice="已完成的任务无需切换 IP";return}
      s.showRoute=false
      s.action("switch-ip",t.ID)
     }
    }),buttonSecondary,buttonMedium)
    styleTactileButton(ui.Button(c,"确定").OnClick(func(){s.showRoute=false}),buttonPrimary,buttonMedium)
   })
  })
 })
}
func (s *downloadUI) restartModal(c *ui.Context){
 ui.Modal(c,&s.showRestart,func(){
  ui.Column(c).Width(420).Gap(12).Children(func(){
   ui.Text(c,"重新下载").Bold().FontSize(18)
   ui.Text(c,"将删除该任务已下载的文件和断点续传记录，从 0% 重新下载。此操作不可撤销。")
   ui.Row(c).Justify(ui.End).Gap(8).Children(func(){
    styleTactileButton(ui.Button(c,"取消").OnClick(func(){s.showRestart=false}),buttonGhost,buttonMedium)
    styleTactileButton(ui.Button(c,"重新下载").OnClick(func(){
     id:=s.restartID
     s.showRestart=false
     s.action("restart",id)
    }),buttonPrimary,buttonMedium)
   })
  })
 })
}

func (s *downloadUI) selectedTask()(Task,bool){
 if s.selected<0||s.selected>=len(s.items){return Task{},false}
 return s.items[s.selected],true
}

func (s *downloadUI) action(kind,id string){
 route:=s.chosenRoute()
 go func(){
  var err error
  switch kind {
  case "stop":err=s.service.Stop(id)
  case "resume":_,err=s.service.Resume(id)
  case "open":err=s.service.Open(id)
  case "folder":err=s.service.Reveal(id)
  case "copy":err=s.service.CopyLink(id)
  case "restart":_,err=s.service.Restart(id)
  case "switch-ip":_,err=s.service.SwitchRoute(id,route)
  case "delete":err=s.service.Remove(id,false)
  case "delete-files":err=s.service.Remove(id,true)
  }
  tasks:=s.service.List()
  s.win.Update(func(){
   s.items=tasks
   if kind=="delete"||kind=="delete-files"{s.selected=-1}
   if err!=nil{s.notice="操作失败："+err.Error()}else{s.notice="操作已执行："+kind}
  })
 }()
}

func (s *downloadUI) browse(){
 name:=s.filename
 if name==""{name="GoogleDrive-file"}
 initial:=filepath.Join(s.folder,name)
 go func(){
  path,err:=mygo.Dialog.Save(mygo.SaveDialogOptions{Parent:s.win,Title:"选择保存文件",DefaultPath:initial})
  s.win.Update(func(){
   if err!=nil{s.notice="无法打开保存窗口："+err.Error();return}
   if path!=""{
    s.folder=filepath.Dir(path)
    s.filename=filepath.Base(path)
   }
  })
 }()
}

func (s *downloadUI) start(){
 raw:=strings.TrimSpace(s.url)
 if _,err:=s.service.Validate(raw);err!=nil{s.notice="仅支持 Google Drive 公开文件链接";return}
 name:=strings.TrimSpace(s.filename)
 if name==""||name=="."||name==".."||filepath.Base(name)!=name||
  strings.ContainsAny(name,"/\\<>:\"|?*"){
  s.notice="请输入有效文件名，并保留正确的扩展名"
  return
 }
 folder:=filepath.Clean(strings.TrimSpace(s.folder))
 if folder==""{s.notice="请选择保存目录";return}
 path:=filepath.Join(folder,name)
 task,err:=s.service.StartRouted(raw,path,s.inspection.Total,s.chosenRoute())
 if err!=nil{s.notice="无法开始下载："+err.Error();return}
 s.showNew=false;s.selected=0
 s.notice="开始下载："+task.Name
 s.items=s.service.List()
}

func (s *downloadUI) askDelete(id string,withFiles bool){
 s.deleteID=id
 s.deleteWithFiles=withFiles
 s.showDelete=true
}

func fileSize(n int64)string{
 if n<0{return "—"}
 if n<1024{return fmt.Sprintf("%d B",n)}
 units:=[]string{"KB","MB","GB","TB"}
 v:=float64(n)
 for _,unit:=range units {
  v/=1024
  if v<1024{return fmt.Sprintf("%.1f %s",v,unit)}
 }
 return fmt.Sprintf("%.1f PB",v/1024)
}

func statusName(state string)string{
 switch state{
 case "running":return "下载中"
 case "stopping":return "正在停止"
 case "complete":return "已完成"
 case "stopped":return "已暂停"
 case "failed":return "失败"
 default:return state
 }
}

func progress(t Task)float64{
 if t.State=="complete"{return 1}
 if t.Total>0{
  return math.Min(0.999,float64(t.Bytes)/float64(t.Total))
 }
 return -1
}

func (s *downloadUI) view(c *ui.Context){
 ui.Column(c).Fill().Padding(12).Gap(9).Children(func(){
  ui.Row(c).Gap(12).Padding(4,6).Children(func(){
   ui.Text(c,"gdirect IDM").Bold().FontSize(19)
   ui.Text(c,"· Native").TextColor(c.Theme().TextMuted)
   ui.Spacer(c)
   ui.Text(c,s.service.Status()).TextColor(c.Theme().TextMuted)
  })
  ui.Divider(c)
  s.toolbar(c)
  ui.Divider(c)
  ui.Row(c).Padding(3,6).Children(func(){
   ui.Text(c,"下载列表").Bold()
   ui.Spacer(c)
   ui.Text(c,fmt.Sprintf("%d 项",len(s.items))).TextColor(c.Theme().TextMuted)
  })
  s.table(c)
  ui.Divider(c)
  s.footer(c)
 })
 s.dialog(c)
 s.deleteDialog(c)
 s.restartModal(c)
 s.routeModal(c)
}

func (s *downloadUI) toolbar(c *ui.Context){
 task,ok:=s.selectedTask()
 running:=ok&&task.State=="running"
 resumable:=ok&&(task.State=="stopped"||task.State=="failed")
 ui.Toolbar(c,func(){
  styleTactileButton(ui.Button(c,"＋ 新建任务").OnClick(func(){s.prompt("")}),buttonPrimary,buttonMedium)
  styleTactileButton(ui.Button(c,"▶ 继续").Disabled(!resumable).OnClick(func(){s.action("resume",task.ID)}),buttonSecondary,buttonMedium)
  styleTactileButton(ui.Button(c,"Ⅱ 暂停").Disabled(!running).OnClick(func(){s.action("stop",task.ID)}),buttonSecondary,buttonMedium)
  styleTactileButton(ui.Button(c,"↻ 重新下载").Disabled(!ok).OnClick(func(){s.restartID=task.ID;s.showRestart=true}),buttonSecondary,buttonMedium)
  styleTactileButton(ui.Button(c,"✕ 删除").Disabled(!ok).OnClick(func(){s.askDelete(task.ID,false)}),buttonTertiary,buttonMedium)
  styleTactileButton(ui.Button(c,"▣ 打开位置").Disabled(!ok).OnClick(func(){s.action("folder",task.ID)}),buttonTertiary,buttonMedium)
  ui.Spacer(c)
  styleTactileButton(ui.Button(c,"线路…").OnClick(func(){s.showRoute=true;s.loadRoutes(false)}),buttonTertiary,buttonMedium)
 }).Label("下载任务操作")
}

func (s *downloadUI) table(c *ui.Context){
 cols:=[]ui.TableColumn{
  {Title:"文件名",Width:250},
  {Title:"大小",Width:100,Align:ui.End},
  {Title:"进度",Width:155},
  {Title:"状态",Width:100},
  {Title:"速度",Width:95,Align:ui.End},
  {Title:"剩余时间",Width:92,Align:ui.End},
  {Title:"保存位置"},
 }
 s.list.Selected=&s.selected
 s.list.Key=func(i int)any{return s.items[i].ID}
 s.list.CellMenu=func(m *ui.Menu,row,col int){
  if row<0||row>=len(s.items){return}
  t:=s.items[row]
  if m.Item("打开文件").Disabled(t.State!="complete").Chosen(){s.action("open",t.ID)}
  if m.Item("打开所在文件夹").Chosen(){s.action("folder",t.ID)}
  if m.Item("复制 Google Drive 链接").Chosen(){s.action("copy",t.ID)}
  m.Separator()
  if m.Item("暂停").Disabled(t.State!="running").Chosen(){s.action("stop",t.ID)}
  if m.Item("继续").Disabled(t.State!="stopped"&&t.State!="failed").Chosen(){s.action("resume",t.ID)}
  if m.Item("重新下载（从头开始）").Chosen(){s.restartID=t.ID;s.showRestart=true}
  if m.Item("切换为当前选定线路并继续").Disabled(t.State=="complete").Chosen(){s.action("switch-ip",t.ID)}
  m.Separator()
  if m.Item("从列表删除").Chosen(){s.askDelete(t.ID,false)}
  if m.Item("删除任务及本地文件").Chosen(){s.askDelete(t.ID,true)}
 }
 if ui.Table(c,&s.list,cols,len(s.items),func(row,col int){
  t:=s.items[row]
  switch col{
  case 0:ui.Text(c,t.Name).SingleLine()
  case 1:
   label:="—"
   if t.Total>0{label=fileSize(t.Total)}else if t.State=="complete"{label=fileSize(t.Bytes)}
   ui.Text(c,label).SingleLine()
  case 2:
   p:=progress(t)
   ui.Row(c).Gap(5).AlignItems(ui.Center).Children(func(){
    // MyGo Progress takes 0..1, not 0..100; fixed track width keeps
    // the visual fill aligned with the displayed percentage in every row.
    ui.Progress(c,p).Width(96).Label("下载进度")
    label:="未知"
    if p>=0{label=fmt.Sprintf("%.0f%%",p*100)}
    ui.Text(c,label).FontSize(11).TextColor(c.Theme().TextMuted).SingleLine()
   })
  case 3:
   state:=statusName(t.State)
   if t.State=="failed"&&t.Error!=""{state+=" · "+t.Error}
   ui.Text(c,state).SingleLine()
  case 4:
   value:="—"
   if t.State=="running"{value=fileSize(int64(t.Speed))+"/s"}
   ui.Text(c,value).SingleLine()
  case 5:
   eta:="—"
   if t.State=="running"&&t.ETASeconds>0{eta=fmt.Sprintf("%d:%02d",t.ETASeconds/60,t.ETASeconds%60)}
   ui.Text(c,eta).SingleLine()
  case 6:ui.Text(c,t.Path).TextColor(c.Theme().TextMuted).SingleLine()
  }
 }).Grow(1).Submitted(){
  if t,ok:=s.selectedTask();ok{s.action("open",t.ID)}
 }
}

func (s *downloadUI) footer(c *ui.Context){
 active:=0
 var speed float64
 for _,t:=range s.items{
  if t.State=="running"{active++;speed+=t.Speed}
 }
 ui.Row(c).Gap(16).Padding(3).Children(func(){
  ui.Text(c,fmt.Sprintf("总任务 %d  ·  下载中 %d",len(s.items),active)).TextColor(c.Theme().TextMuted)
  ui.Text(c,"总速度 "+fileSize(int64(speed))+"/s").TextColor(c.Theme().TextMuted)
  ui.Text(c,"线路："+s.routeChoice).TextColor(c.Theme().TextMuted).SingleLine()
  ui.Spacer(c)
  if s.notice!=""{ui.Text(c,s.notice).SingleLine().TextColor(c.Theme().TextMuted)}
 })
}

func (s *downloadUI) dialog(c *ui.Context){
 ui.Modal(c,&s.showNew,func(){
  ui.Column(c).Width(520).Gap(10).Children(func(){
   ui.Text(c,"新建 Google Drive 下载").Bold().FontSize(18)
   ui.Text(c,"公开文件链接（无需 Google 登录）").TextColor(c.Theme().TextMuted)
   if ui.TextInput(c,&s.url).Placeholder("https://drive.google.com/file/d/...").Changed(){
    if _,err:=s.service.Validate(s.url);err==nil{s.inspect()}
   }
   if s.inspecting {ui.Progress(c,-1).Label("正在获取文件信息")}
   ui.Text(c,"文件名（可重命名）")
   ui.TextInput(c,&s.filename).Placeholder("读取中，或手动输入文件名")
   ui.Text(c,"保存目录")
   ui.Row(c).Gap(8).Children(func(){
    ui.TextInput(c,&s.folder).Grow(1)
    styleTactileButton(ui.Button(c,"浏览…").OnClick(s.browse),buttonSecondary,buttonMedium)
   })
   if s.inspection.Total>0{
    ui.Text(c,"原文件大小："+fileSize(s.inspection.Total)).TextColor(c.Theme().TextMuted)
   }
   if s.notice!=""{ui.Text(c,s.notice).TextColor(c.Theme().TextMuted)}
   ui.Row(c).Gap(10).Justify(ui.End).Children(func(){
    styleTactileButton(ui.Button(c,"取消").OnClick(func(){s.showNew=false;s.inspectionSeq++}),buttonGhost,buttonMedium)
    styleTactileButton(ui.Button(c,"开始下载").OnClick(s.start),buttonPrimary,buttonLarge)
   })
  })
 })
}

func (s *downloadUI) deleteDialog(c *ui.Context){
 ui.Modal(c,&s.showDelete,func(){
  ui.Column(c).Width(380).Gap(12).Children(func(){
   ui.Text(c,"确认删除").Bold().FontSize(18)
   msg:="确定从下载列表删除这项任务？已下载的文件会保留。"
   if s.deleteWithFiles{msg="确定删除任务，并删除本地文件和断点续传数据？此操作无法撤销。"}
   ui.Text(c,msg)
   ui.Row(c).Justify(ui.End).Gap(8).Children(func(){
    styleTactileButton(ui.Button(c,"取消").OnClick(func(){s.showDelete=false}),buttonGhost,buttonMedium)
    styleTactileButton(ui.Button(c,"删除").OnClick(func(){
     id:=s.deleteID
     withFiles:=s.deleteWithFiles
     s.showDelete=false
     if withFiles{s.action("delete-files",id)}else{s.action("delete",id)}
    }),buttonPrimary,buttonMedium)
   })
  })
 })
}
