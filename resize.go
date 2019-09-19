package main

import (
	"fmt"
	"image"
    "image/png"
    "image/color"
    "os"
    //"math/rand"
)

func readImg(fname string) (r, g, b, a [][]uint32, width, height int, err error) {

    file, err := os.Open(fname)
    if err != nil{
        return nil, nil, nil, nil, 0, 0, err
    }
    defer file.Close()
    image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
    img, _, err := image.Decode(file)

    if err != nil {
        return nil, nil, nil, nil, 0, 0, err
    }
    bounds := img.Bounds()
    width, height = bounds.Max.X, bounds.Max.Y

    r = make([][]uint32, height+4)
    g = make([][]uint32, height+4)
    b = make([][]uint32, height+4)
    a = make([][]uint32, height+4)

    for y:=0; y< height+4; y++{
        r[y] = make([]uint32, width+4)
        g[y] = make([]uint32, width+4)
        b[y] = make([]uint32, width+4)
        a[y] = make([]uint32, width+4)
        r[y][0], g[y][0], b[y][0], a[y][0] = img.At(0, y).RGBA()
        r[y][1], g[y][1], b[y][1], a[y][1] = img.At(0, y).RGBA()
        r[y][width], g[y][width], b[y][width], a[y][width] = img.At(width-1, y).RGBA()
        r[y][width+1], g[y][width+1], b[y][width+1], a[y][width+1] = img.At(width-1, y).RGBA()
    }

    for y := 0; y < height; y++ {
        for x := 0; x < width; x++ {
            nr, ng, nb, na := img.At(x, y).RGBA()

            r[y+1][x+1] = nr
            g[y+1][x+1] = ng
            b[y+1][x+1] = nb
            a[y+1][x+1] = na

        }
    }
    return r, g, b, a, width, height, err
}

func cubicInterpolate(line []uint32)uint32{
    var mid uint32 = 0 //rand.Uint32() %2
    //var right uint32 = 3*(line[1] + line[2])/4
    //var left uint32 = (line[1] + line[2])/4
      //  f(x)/2 = ax^3 + bx^2 + cx +d 
    a := uint32((-line[0] + 3*line[1] - 3*line[2] + line[3])/2)
    b := uint32((2*line[0] - 5*line[1] + 4*line[2] - line[3])/2)
    c := uint32((-line[0] + line[2])/2)
    d := line[1]
    f_mid :=uint32((a)*(mid*mid*mid) + (b)*(mid*mid) + (c)*mid + uint32(d))
	//fmt.Println(f_mid)
	
    //f_right :=uint32((a*(right^3) + b*(right^2) + c^right + d)/2) 
    //f_left :=uint32((a*(left^3) + b*(left^2) + c^left + d)/2) 
    ///new_points := []uint32{f_left, f_mid, f_right}
    return f_mid
	//return uint32(newval)
}

func biCubicInterpolate(line [][]uint32)(uint32,uint32, uint32, uint32, uint32){
    //takes in a region of [4][4]uint32 and returns a region of [3][3]uint32 values which have been interpolated from the centre of the image
    col := make([]uint32, 4, 4)
    //row := make([]uint32, 2, 2)
    for i := 0; i < 4; i++{
        col[i] = cubicInterpolate(line[:][i])
    }
    center_point := cubicInterpolate(col)

        //returns the uint32 value for the bicubic interpolation
        //return center, right, left, down, up
    return center_point, col[1], col[2], cubicInterpolate(line[1][:]), cubicInterpolate(line[2][:]) 
}

func interpolateChannel(img [][]uint32, width, height int, id uint32, ch chan map[MapKey]Fullpixel){
    /*arguments:
        img is a [height] * [width] size image of a single color value (e.g. red, green, blue)
        ch is a channel which accepts a map, with a key of type MapKey and returns the fullpixel value
        id is the type of channel
    */
    blank := Fullpixel{}
    for y:= 0; y < height; y++ {
        for x := 0; x < width; x++{
            line := [][]uint32{
                {img[y][x], img[y][x+1],img[y][x+2], img[y][x+3] },
                {img[y+1][x], img[y+1][x+1],img[y+1][x+2], img[y+1][x+3] },
                {img[y+2][x], img[y+2][x+1],img[y+2][x+2], img[y+2][x+3] },
                {img[y+3][x], img[y+3][x+1],img[y+3][x+2], img[y+3][x+3] },
            }
            center, right, left, down, up := biCubicInterpolate(line)
            pixelmap := <- ch
            p := pixelmap[MapKey{x:x, y:y}]
            if p == blank{
                p= Fullpixel{}
            }
            p.pixel_center[id] = center
            p.pixel_right[id] = right
            p.pixel_left[id] = left
            p.pixel_down[id] = down
            p.pixel_up[id] = up
            p.colors_added ++
            pixelmap[MapKey{x:x, y:y}] = p
            ch <- pixelmap
        }
    }
    
}

func paintSection(name string, ch chan map[MapKey]Fullpixel, width, height int, r, g, b, a [][]uint32){
    topleft := image.Point{0, 0}
    bottomright := image.Point{2*width, 2*height}
    imgRGBA := image.NewRGBA(image.Rectangle{topleft, bottomright})


    for y := 0; y < height+4; y++{
        for x := 0; x < width+4; x++{
            col :=  color.RGBA{uint8(r[y][x]), uint8(g[y][x]), uint8(b[y][x]), uint8(a[y][x])}
            imgRGBA.Set(2*x, 2*y-2, col)
        }
    }

    curr_x := 0
    curr_y := 0
    for{
        pixelmap := <- ch
        p := pixelmap[MapKey{x:curr_x, y:curr_y}]
        if p.colors_added == 4 {
            //print the pixel on the screen
            //fmt.Println("addded at", 2*curr_x-1, 2*curr_y-1, "with p = ", p)
            col :=  color.RGBA{uint8(p.pixel_center[0]), uint8(p.pixel_center[1]), uint8(p.pixel_center[2]), uint8(p.pixel_center[3])}
            //imgRGBA.Set(2*curr_x+5, 2*curr_y+5, col) for starting at 1 and incrementing by 2 each time
            //imgRGBA.Set(2*curr_x+3, 2*curr_y+5, col)
            imgRGBA.Set(2*curr_x+2, 2*curr_y+1, col)
            
            col =  color.RGBA{uint8(p.pixel_right[0]), uint8(p.pixel_right[1]), uint8(p.pixel_right[2]), uint8(p.pixel_right[3])}
            imgRGBA.Set(2*curr_x+1, 2*curr_y+1, col)
            col =  color.RGBA{uint8(p.pixel_left[0]), uint8(p.pixel_left[1]), uint8(p.pixel_left[2]), uint8(p.pixel_left[3])}
            imgRGBA.Set(2*curr_x, 2*curr_y+1, col)
            col =  color.RGBA{uint8(p.pixel_up[0]), uint8(p.pixel_left[1]), uint8(p.pixel_left[2]), uint8(p.pixel_left[3])}
            imgRGBA.Set(2*curr_x+1, 2*curr_y+2, col)
            col =  color.RGBA{uint8(p.pixel_down[0]), uint8(p.pixel_down[1]), uint8(p.pixel_down[2]), uint8(p.pixel_down[3])}
            imgRGBA.Set(2*curr_x+1, 2*curr_y, col)
            
            p.colors_added++;
            pixelmap[MapKey{x:curr_x, y:curr_y}] = p
            curr_x += 2
            //advance to next column or row
            if curr_x >= width -1{
                curr_x = 0
                curr_y += 2
            }
            if curr_y >= height -1{
                break
            }
        }
        ch <- pixelmap
    }
    scaled_name := fmt.Sprintf("scaled_%s", name)
    f, _ := os.Create(scaled_name)
    png.Encode(f, imgRGBA)
    
    image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
    file, err := os.Open(scaled_name)
    img, _, err := image.Decode(file)
    topleft = image.Point{0, 0}
    bottomright = image.Point{int(float64(width)*6/4),int(float64(height)*6/4) }
    imgNEW := image.NewRGBA(image.Rectangle{topleft, bottomright})
    if err != nil{
        fmt.Println(err)
        return
        }
    var col [4]uint32
    reduced_y := 0
    for y := 0; y < height*6; y++{
        if  y % 4 == 0 {
            reduced_y ++ 
        }
        reduced_x := 0
        for x:= 0; x <width*6; x++{
            if  x % 4 == 0 {
                reduced_x ++ 
            }
            col[0], col[1],col[2], col[3]= img.At(y, x).RGBA()
            //co := color.RGBA(uint8(col[0]), uint8(col[1]), uint8(col[2]), uint8(col[3])
            imgNEW.Set(y-reduced_y, x-reduced_x, color.RGBA{uint8(col[0]), uint8(col[1]), uint8(col[2]), uint8(col[3])})
        }
    }
    scaled_name_n := fmt.Sprintf("fixed_scaled_%s", name)
    f, _ = os.Create(scaled_name_n)
    png.Encode(f, imgNEW)

}

func scaleOrDeepFry(name string){
    r, g, b, a, width, height, err := readImg(name)
    if err != nil {
        fmt.Println(err)
        return
    }
    ch := make(chan map[MapKey]Fullpixel, 1)
    ch <- map[MapKey]Fullpixel{}
    go interpolateChannel(r, width, height, 0, ch)
    go interpolateChannel(g, width, height, 1, ch)
    go interpolateChannel(b, width, height, 2, ch)
    go interpolateChannel(a, width, height, 3, ch)

    paintSection(name, ch, width, height, r, g, b, a)
}


type MapKey struct{
    x, y int
}

type Fullpixel struct{
    pixel_center [4]uint32
    pixel_up [4]uint32
    pixel_down [4]uint32
    pixel_right [4]uint32
    pixel_left  [4]uint32
    colors_added uint32
}


func main(){
    scaleOrDeepFry("base.png")
    scaleOrDeepFry("test_resizing.png")
    scaleOrDeepFry("small_test.png")
	scaleOrDeepFry("flowers.png")

}

//    "terminal.integrated.shell.windows": "C:\\WINDOWS\\System32\\WindowsPowerShell\\v1.0\\powershell.exe",
