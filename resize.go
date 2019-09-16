package main

import (
	"fmt"
	"image"
    "image/png"
    "image/color"
    "os"
    //"math"
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
    ///
    topleft := image.Point{0, 0}
    bottomright := image.Point{2*width, 2*height}
    imgRGBA := image.NewRGBA(image.Rectangle{topleft, bottomright})
    ///

    //r, g, b, a = [][]uint32{}, [][]uint32{}, [][]uint32{}, [][]uint32{}
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
    //fmt.Println(width+4, height+4, len(r[0]), len(r))
    for y := 0; y < len(r); y++{
        for x := 0; x < len(r[0]); x++{
        ///
        col :=  color.RGBA{uint8(r[y][x]), uint8(g[y][x]), uint8(b[y][x] ), uint8(a[y][x])}
        imgRGBA.Set(x, y, col)
        ///
        }
    }
    f, _ := os.Create("output_real.png")
    png.Encode(f, imgRGBA)
    return r, g, b, a, width, height, err
}

func cubicInterpolate(line []uint32)uint32{
    var center float64 = 0.5
    //var right uint32 = 3*(line[1] + line[2])/4
    //var left uint32 = (line[1] + line[2])/4
      //  f(x)/2 = ax^3 + bx^2 + cx +d 

    if line == nil {
        return 0
    }  
    newval := center*(float64(line[2]) - float64(line[0]) +center*(2*float64(line[0]) - 5*float64(line[1]) + 4*float64(line[2]) - float64(line[3]) + center*(3*(float64(line[1]) - float64(line[2])) + float64(line[3]) - float64(line[0]))))
    newval = uint32(line[1] + newval/2)
    

/*	
    a := float64((-line[0] + 3*line[1] - 3*line[2] + line[3])/2)
    b := float64((2*line[0] - 5*line[1] + 4*line[2] - line[3])/2)
    c := float64((-line[0] + line[2])/2)
    d := line[1]
    f_mid :=uint32(((a)*(math.Pow(mid, 3)) + (b)*(math.Pow(mid, 2)) + (c)*mid + float64(d))/2)
	//fmt.Println(f_mid)
	
    //f_right :=uint32((a*(right^3) + b*(right^2) + c^right + d)/2) 
    //f_left :=uint32((a*(left^3) + b*(left^2) + c^left + d)/2) 
    ///new_points := []uint32{f_left, f_mid, f_right}
    return f_mid*/
	return uint32(newval)
}

func biCubicInterpolate(line [][]uint32)(uint32){
    //takes in a region of [4][4]uint32 and returns a region of [3][3]uint32 values which have been interpolated from the centre of the image
    col := make([]uint32, 4, 4)
    //row := make([]uint32, 2, 2)
    for i := 0; i < 4; i++{
        col[i] = cubicInterpolate(line[:][i])
    }
    center_point := cubicInterpolate(col)
    /*
    for k := 1; k < 3;k++{
        row[k] = cubicInterpolate(line[k][:])
    }
    interpolated := [][]uint32{
        {line[1][1], col[1], line[1][2]},
        {row[0], center_point, row[1]},
        {line[2][1], col[2], line[2][2]},
    }*/

    return center_point //returns the uint32 value for the bicubic interpolation
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
            new_point := biCubicInterpolate(line)
            pixelmap := <- ch
            p := pixelmap[MapKey{x:x, y:y}]
            if p == blank{
                p= Fullpixel{}
            }
            p.colortype[id] = new_point
            p.colors_added ++
            pixelmap[MapKey{x:x, y:y}] = p
            ch <- pixelmap
        }
    }
    
}

func paintSection(ch chan map[MapKey]Fullpixel, width, height int, r, g, b, a [][]uint32){
    topleft := image.Point{0, 0}
    bottomright := image.Point{2*width, 2*height}
    imgRGBA := image.NewRGBA(image.Rectangle{topleft, bottomright})


	
    for y := 0; y < height+4; y++{
        for x := 0; x < width+4; x++{
            col :=  color.RGBA{uint8(r[y][x]), uint8(g[y][x]), uint8(b[y][x]), uint8(a[y][x])}
            imgRGBA.Set(2*x, 2*y, col)
        }
    }
    /*
    f, _ := os.Create("output.png")
    png.Encode(f, imgRGBA)
    */
    curr_x := 0
    curr_y := 0
    for{
        pixelmap := <- ch
        p := pixelmap[MapKey{x:curr_x, y:curr_y}]
        if p.colors_added == 4 {
            //print the pixel on the screen
            col :=  color.RGBA{uint8(p.colortype[0]), uint8(p.colortype[1]), uint8(p.colortype[2]), uint8(p.colortype[3])}
            //imgRGBA.Set(2*curr_x+5, 2*curr_y+5, col) for starting at 1 and incrementing by 2 each time
			imgRGBA.Set(2*curr_x+1, 2*curr_y+1, col)
            curr_x +=1
            //advance to next column or row
            if curr_x == width -1{
                curr_x = 1
                curr_y += 1
            }
            if curr_y == height -1{
                break
            }
        }

        ch <- pixelmap
    }
    f, _ := os.Create("scaled_output.png")
    png.Encode(f, imgRGBA)
}

type MapKey struct{
    x, y int
}

type Fullpixel struct{
    colortype [4]uint32
    colors_added uint32
}

func main(){
    r, g, b, a, width, height, err := readImg("base.PNG")
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

    paintSection(ch, width, height, r, g, b, a)




}
