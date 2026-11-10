package driveurl
import "testing"
func TestParse(t *testing.T) {
 tests:=[]struct{url string; valid bool; folder bool}{
 {"https://drive.google.com/file/d/1aBcD_efghijkLMN/view?usp=sharing",true,false},
 {"https://drive.google.com/open?id=1aBcD_efghijkLMN&resourcekey=0-abc",true,false},
 {"https://drive.google.com/uc?export=download&id=1aBcD_efghijkLMN",true,false},
 {"https://drive.usercontent.google.com/download?id=1aBcD_efghijkLMN",true,false},
 {"https://drive.google.com/drive/folders/1aBcD_efghijkLMN",false,true},
 {"https://drive.google.com.evil.com/file/d/1aBcD_efghijkLMN/view",false,false},
 {"http://drive.google.com/file/d/1aBcD_efghijkLMN/view",false,false},
 {"https://docs.google.com/document/d/1aBcD_efghijkLMN/edit",false,false},
 {"https://drive.google.com/file/d/1aBcD_efghijkLMN/view/extra",false,false},
 }
 for _,tt:=range tests{
   file,err:=Parse(tt.url)
   if (err==nil)!=tt.valid {t.Errorf("Parse(%s) = %v, %v",tt.url,file,err)}
   if tt.folder&&err!=ErrFolder{t.Errorf("folder error for %s: %v",tt.url,err)}
 }
}
