package main

import(
	"resize"		//place the folder called 'resize' in the \src\ folder (e.g. C:\Go\src if using windows)
)

func main(){
	//call the Resize function from the package to resize the image to twice the resolution
	resize.Resize("the_expanse.png")		
}