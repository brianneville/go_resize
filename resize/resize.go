package resize
/*
    By Brian Neville (https://github.com/tropical-bn)
    This is an image resizing tool that I wrote - also my first project in Golang (hyyype!). 
    The function 'Resize' takes in the filename of a png image, and will create a version of 
    that file with the name 'scaled_<filename>.png' 
*/

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

func interpolateChannel(img [][]uint32, width, height int, id uint32, ch chan map[mapKey]fullpixel){
    /*arguments:
        img is a [height] * [width] size image of a single color value (e.g. red, green, blue)
        ch is a channel which accepts a map, with a key of type MapKey and returns the fullpixel value
        id is the type of channel
    */
    blank := fullpixel{}
    for y:= 0; y < height; y++ {
        for x := 0; x < width; x++{

            center := img[y+1][x+1]
            pixelmap := <- ch
            p := pixelmap[mapKey{x:x, y:y}]
            if p == blank{
                p= fullpixel{}
            }
            p.pixel_center[id] = center
            p.colors_added ++
            pixelmap[mapKey{x:x, y:y}] = p
            ch <- pixelmap      //push into channel to be drawn or to have other colors added to pixel
        } 
    }
    
}

func paintImage(name string, ch chan map[mapKey]fullpixel, width, height int, r, g, b, a [][]uint32){
    topleft := image.Point{0, 0}
    bottomright := image.Point{2*width, 2*height}
    imgRGBA := image.NewRGBA(image.Rectangle{topleft, bottomright})

    //draw the original image with gaps to be interpolated
    for y := 0; y < height+4; y++{
        for x := 0; x < width+4; x++{
            col :=  color.RGBA{uint8(r[y][x]), uint8(g[y][x]), uint8(b[y][x]), uint8(a[y][x])}
            imgRGBA.Set(2*x, 2*y-2, col)
            imgRGBA.Set(2*x-1, 2*y-2, col)

        }
    }

    //interpolate the rest 
    curr_x := 0
    curr_y := 0
    for{
        pixelmap := <- ch       //pull from channel
        p := pixelmap[mapKey{x:curr_x, y:curr_y}]
        if p.colors_added == 4 {
            //print the pixels on the screen
            col :=  color.RGBA{uint8(p.pixel_center[0]), uint8(p.pixel_center[1]), uint8(p.pixel_center[2]), uint8(p.pixel_center[3])}
            imgRGBA.Set(2*curr_x+2, 2*curr_y+1, col)
            imgRGBA.Set(2*curr_x+1, 2*curr_y+1, col)
            imgRGBA.Set(2*curr_x+3, 2*curr_y+1, col)       

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

        ch <- pixelmap      //push to channel
    }

    //do one last pass to correct left hand side errors
    for p := 1; p < height*2; p+=2{
        t0, t1, t2, t3 := imgRGBA.At(0, p-1).RGBA()
        
        b0, b1, b2, b3 := imgRGBA.At(0, p+1).RGBA()
        r0, r1, r2, r3 := imgRGBA.At(1, p).RGBA()
        c0, c1, c2, c3 := uint32(t0/3 + b0/3 + r0/3), uint32(t1/3 + b1/3 + r1/3), uint32(t2/3 + b2/3 + r2/3),uint32(t3/3 + b3/3 + r3/3)
        
      imgRGBA.Set(0, p, color.RGBA{uint8(c0), uint8(c1), uint8(c2), uint8(c3)})
    }
    
    for p := 0; p < height*2-1; p++{
        c0, c1, c2, c3 := imgRGBA.At(1, p).RGBA()
        imgRGBA.Set(0, p, color.RGBA{uint8(c0), uint8(c1), uint8(c2), uint8(c3)})

    }
    
    scaled_name := fmt.Sprintf("scaled_%s", name)
    f, _ := os.Create(scaled_name)
    png.Encode(f, imgRGBA)
}

func Resize(name string){
    r, g, b, a, width, height, err := readImg(name) //split original image into r, g, b, a layers
    if err != nil {
        fmt.Println(err)
        return
    }
    ch := make(chan map[mapKey]fullpixel, 1)    
    ch <- map[mapKey]fullpixel{}        
    // ^inialise empty dictionary. this will be used to pass data between painter and interpolator

    //run in goroutines!
    go interpolateChannel(r, width, height, 0, ch)
    go interpolateChannel(g, width, height, 1, ch)
    go interpolateChannel(b, width, height, 2, ch)
    go interpolateChannel(a, width, height, 3, ch)

    paintImage(name, ch, width, height, r, g, b, a)
}


type mapKey struct{
    x, y int
}

type fullpixel struct{
    pixel_center [4]uint32
    colors_added uint32
}
