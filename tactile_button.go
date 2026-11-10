package main

import (
 "time"

 "github.com/egoist/mygo/ui"
)

// Material styles are applied to MyGo's existing accessible Button element.
// No web components, new button engine, or keyboard handling are introduced.
type buttonVariant string
const (
 buttonPrimary buttonVariant = "primary"
 buttonSecondary buttonVariant = "secondary"
 buttonTertiary buttonVariant = "tertiary"
 buttonGhost buttonVariant = "ghost"
 buttonText buttonVariant = "text"
)
type buttonSize string
const (
 buttonMedium buttonSize = "md"
 buttonLarge buttonSize = "lg"
)

type buttonPalette struct{
 top, bottom ui.Color
 hoverTop, hoverBottom ui.Color
 edge, label ui.Color
 topLight, bottomShade float32
 contact, lift float32
}

func buttonColors(variant buttonVariant) buttonPalette {
 switch variant {
 case buttonPrimary:
  return buttonPalette{
   top:ui.Hex("#404859"),bottom:ui.Hex("#273042"),
   hoverTop:ui.Hex("#4C5566"),hoverBottom:ui.Hex("#313B4C"),
   edge:ui.Hex("#12161F"),label:ui.Hex("#FFFFFF"),
   topLight:0.18,bottomShade:0.22,contact:0.18,lift:0.28,
  }
 case buttonTertiary:
  return buttonPalette{
   top:ui.Hex("#F3F4F6"),bottom:ui.Hex("#E7E9EC"),
   hoverTop:ui.Hex("#FFFFFF"),hoverBottom:ui.Hex("#EAECF0"),
   edge:ui.RGBA(15,23,42,0.10),label:ui.Hex("#313946"),
   topLight:0.85,bottomShade:0.06,contact:0.06,lift:0.08,
  }
 default:
  return buttonPalette{
   top:ui.Hex("#FFFFFF"),bottom:ui.Hex("#F4F5F7"),
   hoverTop:ui.Hex("#FFFFFF"),hoverBottom:ui.Hex("#F8F9FA"),
   edge:ui.RGBA(15,23,42,0.14),label:ui.Hex("#313946"),
   topLight:0.95,bottomShade:0.06,contact:0.06,lift:0.08,
  }
 }
}

// styleTactileButton keeps the original MyGo ui.Button's click, focus,
// disabled, keyboard and accessibility behavior while setting its appearance.
func styleTactileButton(b ui.Element, variant buttonVariant, size buttonSize) ui.Element {
 u:=float32(1)
 labelSize:=float32(14)
 if size==buttonLarge{u=16.0/14.0;labelSize=16}
 b.Height(40*u).Padding(0,18*u).Radius(10*u).
  FontSize(labelSize).FontWeight(600).Cursor(ui.CursorPointer).
  Transition(ui.ElementTransition{Colors:true,Shadows:true,Duration:140*time.Millisecond})
 p:=buttonColors(variant)
 hover:=b.Hovered()
 pressed:=b.Pressed()
 if variant==buttonGhost||variant==buttonText{
  b.NoShadow().Border(0,ui.Transparent).TextColor(ui.Hex("#596170"))
  if variant==buttonGhost{
   b.Background(ui.Transparent)
   if hover{b.Background(ui.Hex("#F3F4F6"))}
   if pressed{
    b.Background(ui.Hex("#E5E7EB"))
    b.InsetShadow(0,1*u,2*u,0,ui.RGBA(15,23,42,0.12))
   }
  }else{
   b.Background(ui.Transparent)
   if hover{b.TextColor(ui.Hex("#252D3B"))}
  }
  if pressed{b.Translate(0,b.Animate("press",u,120*time.Millisecond))}else{b.Translate(0,b.Animate("press",0,120*time.Millisecond))}
  return b
 }
 top,bottom:=p.top,p.bottom
 if hover{top,bottom=p.hoverTop,p.hoverBottom}
 if pressed{top,bottom=p.bottom,p.top}
 b.NoShadow().Border(1*u,p.edge).
  Gradient(top,bottom,180).TextColor(p.label)
 if variant==buttonPrimary{b.TextShadow(0,-u,0,ui.RGBA(0,0,0,0.28))}

 if pressed{
  // Same two inner and two outer material layers as the resting state.
  shade:=float32(0.14)
  if variant==buttonPrimary{shade=0.40}
  b.InsetShadow(0,1*u,2*u,0,ui.RGBA(0,0,0,shade)).
   InsetShadow(0,-1*u,0,0,ui.RGBA(255,255,255,0.90)).
   Shadow(0,1*u,1*u,0,ui.RGBA(0,0,0,p.contact/2)).
   Shadow(0,0,0,0,ui.Transparent)
 }else{
  b.InsetShadow(0,1*u,0,0,ui.RGBA(255,255,255,p.topLight)).
   InsetShadow(0,-1*u,0,0,ui.RGBA(0,0,0,p.bottomShade)).
   Shadow(0,1*u,2*u,0,ui.RGBA(0,0,0,p.contact))
  if hover{
   b.Shadow(0,8*u,16*u,-5*u,ui.RGBA(0,0,0,p.lift))
  }else{
   b.Shadow(0,4*u,10*u,-3*u,ui.RGBA(0,0,0,p.lift))
  }
 }
 pressY:=float32(0)
 if pressed{pressY=u}
 b.Translate(0,b.Animate("press",pressY,120*time.Millisecond))
 return b
}
