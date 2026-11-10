package main

import (
 _ "embed"
 "log"
 "time"

 "github.com/egoist/mygo"
 "github.com/egoist/mygo/ui"
)

//go:embed resources/tray.png
var trayImage []byte

func main(){
 var quitting bool
 mygo.App.OnBeforeQuit(func(*mygo.QuitEvent){quitting=true})
 mygo.App.WhenReady(func(){
  service:=NewService()
  screen:=newDownloadUI(service)
  win:=mygo.NewWindow(mygo.WindowOptions{
   Title:"gdirect IDM",Width:1100,Height:680,MinWidth:780,MinHeight:490,
   StateKey:"main",Content:ui.View(screen.view),
  })
  screen.win=win
  screen.loadRoutes(false)
  menu:=mygo.NewMenu([]*mygo.MenuItem{
   {Role:mygo.RoleAppMenu},
   {Label:"任务",Submenu:[]*mygo.MenuItem{
    {Label:"新建 Google Drive 下载",Accelerator:"CmdOrCtrl+N",Click:func(*mygo.MenuItem,*mygo.Window){win.Update(func(){screen.prompt("")})}},
    {Label:"从剪贴板添加",Accelerator:"CmdOrCtrl+V",Click:func(*mygo.MenuItem,*mygo.Window){win.Update(func(){screen.prompt(mygo.Clipboard.ReadText())})}},
    mygo.Separator(),
    {Label:"关闭窗口（转到托盘）",Role:mygo.RoleClose},
    {Role:mygo.RoleQuit},
   }},
   {Role:mygo.RoleEditMenu},
   {Role:mygo.RoleWindowMenu},
  })
  mygo.App.SetMenu(menu)
  tray,err:=mygo.NewTray(mygo.TrayOptions{
   ID:"gdiretc-idm",Icon:trayImage,ToolTip:"gdirect IDM · Google Drive downloads",
   Menu:mygo.NewMenu([]*mygo.MenuItem{
    {Label:"显示下载管理器",Click:func(*mygo.MenuItem,*mygo.Window){win.Show();win.Focus()}},
    {Label:"新建下载",Click:func(*mygo.MenuItem,*mygo.Window){win.Show();win.Focus();win.Update(func(){screen.prompt("")})}},
    mygo.Separator(),
    {Role:mygo.RoleQuit},
   }),
  })
  if err!=nil{log.Printf("tray unavailable: %v",err)}else{screen.tray=tray}
  win.OnClose(func(e *mygo.CloseEvent){
   if !quitting && screen.tray!=nil {
    e.PreventDefault()
    win.Hide()
   }
  })
  screen.lastClipboard=mygo.Clipboard.ReadText()
  go screen.poll()
 })
 if err:=mygo.App.Run();err!=nil{log.Fatal(err)}
}

func (s *downloadUI) poll(){
 t:=time.NewTicker(550*time.Millisecond)
 defer t.Stop()
 for range t.C{
  tasks:=s.service.List()
  clip:=mygo.Clipboard.ReadText()
  s.win.Update(func(){
   s.items=tasks
   if !s.showNew&&clip!=s.lastClipboard{
    s.lastClipboard=clip
    if _,err:=s.service.Validate(clip);err==nil{s.prompt(clip)}
   }else if clip!=s.lastClipboard {s.lastClipboard=clip}
   if s.selected>=len(s.items){s.selected=-1}
  })
 }
}
