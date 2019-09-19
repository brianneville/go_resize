package main

import (
	"fmt"
	"image"
    "image/png"
    "image/color"
    "os"
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

func two_dimInterpolate(line [][]uint32)(uint32,uint32, uint32, uint32, uint32){
    col := make([]uint32, 2, 2)
    for i := 0; i < 2; i++{
        col[i] = line[0][i]
    }
    center_point := col[0]

    return center_point, col[0], col[1], line[0][0], line[1][0] 
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
                {img[y+1][x+1],img[y+1][x+2] },
                {img[y+2][x+1],img[y+2][x+2]},
            }
            center, right, left, down, up := two_dimInterpolate(line)
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
            imgRGBA.Set(2*x-1, 2*y-2, col)

        }
    }

    curr_x := 0
    curr_y := 0
    for{
        pixelmap := <- ch
        p := pixelmap[MapKey{x:curr_x, y:curr_y}]
        if p.colors_added == 4 {
            //print the pixel on the screen
            col :=  color.RGBA{uint8(p.pixel_center[0]), uint8(p.pixel_center[1]), uint8(p.pixel_center[2]), uint8(p.pixel_center[3])}
            imgRGBA.Set(2*curr_x+2, 2*curr_y+1, col)
            //col :=  color.RGBA{uint8(p.pixel_center[0]), uint8(p.pixel_center[1]), uint8(p.pixel_center[2]), uint8(p.pixel_center[3])}
            imgRGBA.Set(2*curr_x+1, 2*curr_y+1, col)
            imgRGBA.Set(2*curr_x+3, 2*curr_y+1, col)       
           /* 
            col =  color.RGBA{uint8(p.pixel_right[0]), uint8(p.pixel_right[1]), uint8(p.pixel_right[2]), uint8(p.pixel_right[3])}
            imgRGBA.Set(2*curr_x-1, 2*curr_y+1, col)
            col =  color.RGBA{uint8(p.pixel_left[0]), uint8(p.pixel_left[1]), uint8(p.pixel_left[2]), uint8(p.pixel_left[3])}
            imgRGBA.Set(2*curr_x-2, 2*curr_y+1, col)
            col =  color.RGBA{uint8(p.pixel_up[0]), uint8(p.pixel_left[1]), uint8(p.pixel_left[2]), uint8(p.pixel_left[3])}
            imgRGBA.Set(2*curr_x-1, 2*curr_y+2, col)
            col =  color.RGBA{uint8(p.pixel_down[0]), uint8(p.pixel_down[1]), uint8(p.pixel_down[2]), uint8(p.pixel_down[3])}
            imgRGBA.Set(2*curr_x-1, 2*curr_y, col)
            */
            p.colors_added++;
            pixelmap[MapKey{x:curr_x, y:curr_y}] = p
            curr_x += 1
            //advance to next column or row
            if curr_x >= width -1{
                curr_x = 0
                curr_y += 1
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
}

func resize(name string){
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
    /*
    resize("base.png")
    resize("test_resizing.png")
    resize("small_test.png")
	resize("flowers.png")
    resize("rabbits.png")
    */
    resize("test_resizing.png")
}
