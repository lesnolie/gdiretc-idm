package main

import (
  "context"
  "errors"
  "fmt"
  "log"
  "os"
  "os/exec"
  "path/filepath"
  "runtime"
  "sort"
  "strings"
  "sync"
  "time"

  "github.com/egoist/mygo"
  "github.com/lesnolie/gdiretc-idm/internal/driveurl"
)

type Task struct {
  ID string `json:"id"`
  URL string `json:"url"`
  Name string `json:"name"`
  Path string `json:"path"`
  State string `json:"state"`
  Error string `json:"error,omitempty"`
  Bytes int64 `json:"bytes"`
  Started string `json:"started"`
}

type job struct { task Task; cancel context.CancelFunc }
type Service struct {
  mu sync.Mutex
  tasks map[string]*job
  helper string
  helperErr error
}
func NewService() *Service {
  s := &Service{tasks: make(map[string]*job)}
  suffix := ""; if runtime.GOOS == "windows" { suffix = ".exe" }
  if override := os.Getenv("GDIRECT_IDM_HELPER"); override != "" {
    s.helper, s.helperErr = filepath.Abs(override)
  } else {
    resourceDir, err := mygo.App.Path(mygo.PathResources)
    if err == nil { s.helper = filepath.Join(resourceDir, "bin", "gdirect"+suffix) } else { s.helperErr = err }
  }
  if s.helperErr == nil {
    st, err := os.Stat(s.helper)
    if err != nil { s.helperErr = fmt.Errorf("bundled gdirect helper is unavailable: %w", err) }
    if err == nil && st.IsDir() { s.helperErr = errors.New("gdirect helper is a directory") }
  }
  return s
}
func (s *Service) Status() string { if s.helperErr != nil { return s.helperErr.Error() }; return "Ready — public files only" }
func (s *Service) ClipboardText() string { return mygo.Clipboard.ReadText() }
func (s *Service) Validate(raw string) (string,error) {
  file,err := driveurl.Parse(strings.TrimSpace(raw))
  if err != nil { return "",err }
  return file.ID,nil
}
func (s *Service) ChoosePath(ctx context.Context, preferredName string) (string,error) {
  preferredName = filepath.Base(preferredName)
  if preferredName == "" || preferredName == "." || preferredName == string(filepath.Separator) { preferredName = "download.bin" }
  return mygo.Dialog.Save(mygo.SaveDialogOptions{Parent:mygo.CallerWindow(ctx), Title:"Save Google Drive file", DefaultPath:preferredName})
}
func (s *Service) Start(url, path string) (Task,error) {
  if s.helperErr != nil { return Task{},s.helperErr }
  if _,err := driveurl.Parse(url); err != nil { return Task{},err }
  if !filepath.IsAbs(path) { return Task{},errors.New("select an absolute save path") }
  path = filepath.Clean(path)
  if st,err:=os.Stat(filepath.Dir(path)); err!=nil || !st.IsDir() { return Task{},errors.New("save folder does not exist") }
  if _,err:=os.Stat(path); err==nil { return Task{},errors.New("destination exists; choose another name") } else if !errors.Is(err,os.ErrNotExist) { return Task{},err }
  s.mu.Lock(); defer s.mu.Unlock()
  for _,j:=range s.tasks { if j.task.Path==path && j.task.State=="running" {return Task{},errors.New("a task already uses this path")} }
  id:=fmt.Sprint(time.Now().UnixNano())
  ctx,cancel:=context.WithCancel(context.Background())
  t:=Task{ID:id,URL:url,Name:filepath.Base(path),Path:path,State:"running",Started:time.Now().Format(time.RFC3339)}
  s.tasks[id]=&job{task:t,cancel:cancel}
  go s.run(ctx,id)
  return t,nil
}
func (s *Service) run(ctx context.Context,id string) {
  s.mu.Lock(); j:=s.tasks[id]; if j==nil {s.mu.Unlock();return}; url,path:=j.task.URL,j.task.Path; s.mu.Unlock()
  cmd:=exec.CommandContext(ctx,s.helper,"download",url,"-o",path)
  var out limitedBuffer
  cmd.Stdout=&out; cmd.Stderr=&out
  err:=cmd.Run()
  s.mu.Lock(); defer s.mu.Unlock()
  j=s.tasks[id]; if j==nil{return}
  if ctx.Err()!=nil {j.task.State="stopped"} else if err!=nil {
    j.task.State="failed"
    msg:=strings.TrimSpace(out.String())
    if len(msg)>300 {msg=msg[len(msg)-300:]}
    j.task.Error=fmt.Sprintf("%v: %s",err,msg)
  } else {j.task.State="complete"}
  j.cancel=nil
  if st,err:=os.Stat(path);err==nil {j.task.Bytes=st.Size()}
}
type limitedBuffer struct { contents []byte }
func (b *limitedBuffer) Write(data []byte) (int,error) {
  n:=len(data); b.contents=append(b.contents,data...)
  if len(b.contents)>4096 {b.contents=b.contents[len(b.contents)-4096:]}
  return n,nil
}
func (b *limitedBuffer) String() string {return string(b.contents)}
func (s *Service) List() []Task {
  s.mu.Lock(); defer s.mu.Unlock()
  all:=make([]Task,0,len(s.tasks))
  for _,j:=range s.tasks {
    t:=j.task
    if t.State=="running" {if st,err:=os.Stat(t.Path);err==nil {t.Bytes=st.Size()}}
    all=append(all,t)
  }
  sort.Slice(all,func(i,j int)bool{return all[i].Started>all[j].Started})
  return all
}
func (s *Service) Stop(id string) error {
  s.mu.Lock();defer s.mu.Unlock()
  j:=s.tasks[id];if j==nil {return errors.New("unknown download")}
  if j.task.State!="running" || j.cancel==nil{return errors.New("task is not running")}
  j.cancel();return nil
}
func (s *Service) Resume(id string) (Task,error) {
  if s.helperErr!=nil {return Task{},s.helperErr}
  s.mu.Lock();defer s.mu.Unlock()
  j:=s.tasks[id];if j==nil{return Task{},errors.New("unknown download")}
  if j.task.State!="stopped" && j.task.State!="failed" {return Task{},errors.New("only stopped or failed downloads can resume")}
  ctx,cancel:=context.WithCancel(context.Background())
  j.cancel=cancel;j.task.State="running";j.task.Error="";j.task.Started=time.Now().Format(time.RFC3339)
  go s.run(ctx,id)
  return j.task,nil
}
func (s *Service) ShowInFolder(path string) error {
  if _,err:=os.Stat(path);err!=nil{return err}
  mygo.Shell.ShowItemInFolder(path)
  return nil
}
func main() {
  mygo.Bind(NewService())
  mygo.App.WhenReady(func(){
    mygo.NewWindow(mygo.WindowOptions{Title:"gdirect IDM",URL:"/",Width:960,Height:640,MinWidth:690,MinHeight:480,StateKey:"main"})
  })
  if err:=mygo.App.Run();err!=nil{log.Fatal(err)}
}
