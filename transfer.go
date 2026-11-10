package main

import (
 "context"
 "errors"
 "fmt"
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
 ID string
 URL string
 Name string
 Path string
 State string
 Error string
 Bytes int64
 Total int64
 Speed float64
 ETASeconds int64
 Started time.Time
 Finished time.Time
}

type job struct {
 task Task
 cancel context.CancelFunc
 done chan struct{}
 prevBytes int64
 prevTime time.Time
}

type Service struct {
 mu sync.Mutex
 tasks map[string]*job
 helper string
 helperErr error
}

func NewService() *Service {
 s:=&Service{tasks:make(map[string]*job)}
 suffix:=""
 if runtime.GOOS=="windows"{suffix=".exe"}
 if override:=os.Getenv("GDIRECT_IDM_HELPER");override!=""{
  s.helper,s.helperErr=filepath.Abs(override)
 }else{
  var dir string
  dir,s.helperErr=mygo.App.Path(mygo.PathResources)
  if s.helperErr==nil{s.helper=filepath.Join(dir,"bin","gdirect"+suffix)}
 }
 if s.helperErr==nil{
  var st os.FileInfo
  st,s.helperErr=os.Stat(s.helper)
  if s.helperErr==nil&&st.IsDir(){s.helperErr=errors.New("gdirect helper is not an executable")}
 }
 return s
}

func (s *Service) Status() string {
 if s.helperErr!=nil{return "gdirect unavailable: "+s.helperErr.Error()}
 return "Ready · Google Drive public files"
}

func (s *Service) Start(raw,dest string,total int64)(Task,error){
 if s.helperErr!=nil{return Task{},s.helperErr}
 if _,err:=driveurl.Parse(strings.TrimSpace(raw));err!=nil{return Task{},err}
 if !filepath.IsAbs(dest){return Task{},errors.New("choose an absolute save path")}
 dest=filepath.Clean(dest)
 if dest=="."||filepath.Base(dest)=="."{return Task{},errors.New("invalid destination")}
 st,err:=os.Stat(filepath.Dir(dest))
 if err!=nil||!st.IsDir(){return Task{},errors.New("destination folder does not exist")}
 if _,err:=os.Lstat(dest);err==nil{return Task{},errors.New("destination already exists")}
 if !errors.Is(err,os.ErrNotExist){return Task{},err}
 s.mu.Lock()
 defer s.mu.Unlock()
 for _,j:=range s.tasks{
  if filepath.Clean(j.task.Path)==dest{return Task{},errors.New("a task already uses this path")}
 }
 id:=fmt.Sprintf("%d",time.Now().UnixNano())
 ctx,cancel:=context.WithCancel(context.Background())
 now:=time.Now()
 j:=&job{task:Task{ID:id,URL:strings.TrimSpace(raw),Path:dest,Name:filepath.Base(dest),State:"running",Total:total,Started:now},cancel:cancel,done:make(chan struct{}),prevTime:now}
 s.tasks[id]=j
 go s.run(ctx,id,j)
 return j.task,nil
}

func (s *Service) run(ctx context.Context,id string,j *job){
 defer close(j.done)
 cmd:=exec.CommandContext(ctx,s.helper,"download",j.task.URL,"-o",j.task.Path)
 hideHelperWindow(cmd)
 var output limitedBuffer
 cmd.Stdout=&output
 cmd.Stderr=&output
 err:=cmd.Run()
 s.mu.Lock()
 defer s.mu.Unlock()
 current:=s.tasks[id]
 if current!=j{return}
 if stat,e:=os.Stat(j.task.Path);e==nil{j.task.Bytes=stat.Size()}
 j.task.Speed=0
 j.task.ETASeconds=0
 j.task.Finished=time.Now()
 if ctx.Err()!=nil{
  j.task.State="stopped"
 }else if err!=nil{
  j.task.State="failed"
  j.task.Error=fmt.Sprintf("%v: %s",err,strings.TrimSpace(output.String()))
 }else{
  j.task.State="complete"
  if j.task.Total<=0{j.task.Total=j.task.Bytes}
 }
 j.cancel=nil
}

type limitedBuffer struct{b []byte}
func (b *limitedBuffer) Write(p []byte)(int,error){
 b.b=append(b.b,p...)
 if len(b.b)>4096{b.b=b.b[len(b.b)-4096:]}
 return len(p),nil
}
func (b *limitedBuffer) String()string{return string(b.b)}

func (s *Service) List()[]Task{
 s.mu.Lock()
 defer s.mu.Unlock()
 now:=time.Now()
 all:=make([]Task,0,len(s.tasks))
 for _,j:=range s.tasks{
  t:=&j.task
  if t.State=="running"||t.State=="stopping"{
   if st,err:=os.Stat(t.Path);err==nil{
    n:=st.Size()
    if n<t.Bytes{n=t.Bytes}
    t.Bytes=n
   }
   elapsed:=now.Sub(j.prevTime).Seconds()
   if elapsed>=0.6{
    instant:=float64(t.Bytes-j.prevBytes)/elapsed
    if instant<0{instant=0}
    if j.prevTime.Equal(t.Started){t.Speed=instant}else{t.Speed=t.Speed*0.55+instant*0.45}
    j.prevTime=now
    j.prevBytes=t.Bytes
    if t.Total>t.Bytes&&t.Speed>1024{t.ETASeconds=int64(float64(t.Total-t.Bytes)/t.Speed)}else{t.ETASeconds=0}
   }
  }
  all=append(all,*t)
 }
 sort.Slice(all,func(i,k int)bool{return all[i].Started.After(all[k].Started)})
 return all
}
func (s *Service) Stop(id string)error{
 s.mu.Lock();defer s.mu.Unlock()
 j:=s.tasks[id]
 if j==nil{return errors.New("download not found")}
 if j.task.State!="running"||j.cancel==nil{return errors.New("download is not running")}
 j.task.State="stopping"
 j.cancel()
 return nil
}
func (s *Service) Resume(id string)(Task,error){
 if s.helperErr!=nil{return Task{},s.helperErr}
 s.mu.Lock();defer s.mu.Unlock()
 j:=s.tasks[id]
 if j==nil{return Task{},errors.New("download not found")}
 if j.task.State!="stopped"&&j.task.State!="failed"{return Task{},errors.New("only stopped or failed downloads may resume")}
 if j.done!=nil{
  select{case <-j.done:default:return Task{},errors.New("previous download is still stopping")}
 }
 ctx,cancel:=context.WithCancel(context.Background())
 j.task.State="running";j.task.Error="";j.task.Finished=time.Time{}
 j.cancel=cancel;j.done=make(chan struct{})
 j.prevTime=time.Now();j.prevBytes=j.task.Bytes
 go s.run(ctx,id,j)
 return j.task,nil
}
func (s *Service) Remove(id string,deleteFiles bool)error{
 s.mu.Lock()
 j:=s.tasks[id]
 if j==nil{s.mu.Unlock();return errors.New("download not found")}
 if j.cancel!=nil{j.cancel()}
 done:=j.done
 s.mu.Unlock()
 if done!=nil{
  select{case <-done:case <-time.After(12*time.Second):return errors.New("download did not stop; file was not deleted")}
 }
 s.mu.Lock()
 if s.tasks[id]!=j{s.mu.Unlock();return errors.New("download changed while removing")}
 delete(s.tasks,id)
 s.mu.Unlock()
 if deleteFiles{
  if err:=os.Remove(j.task.Path);err!=nil&&!errors.Is(err,os.ErrNotExist){return err}
  if err:=os.Remove(j.task.Path+".aria2");err!=nil&&!errors.Is(err,os.ErrNotExist){return err}
 }
 return nil
}
func (s *Service) Reveal(id string)error{
 s.mu.Lock();j:=s.tasks[id];s.mu.Unlock()
 if j==nil{return errors.New("download not found")}
 if _,err:=os.Stat(j.task.Path);err==nil{mygo.Shell.ShowItemInFolder(j.task.Path);return nil}
 return mygo.Shell.OpenPath(filepath.Dir(j.task.Path))
}
func (s *Service) Open(id string)error{
 s.mu.Lock();j:=s.tasks[id];s.mu.Unlock()
 if j==nil{return errors.New("download not found")}
 if j.task.State!="complete"{return errors.New("file has not finished downloading")}
 if _,err:=os.Stat(j.task.Path);err!=nil{return err}
 return mygo.Shell.OpenPath(j.task.Path)
}
func (s *Service) CopyLink(id string)error{
 s.mu.Lock();j:=s.tasks[id];s.mu.Unlock()
 if j==nil{return errors.New("download not found")}
 mygo.Clipboard.WriteText(j.task.URL)
 return nil
}
