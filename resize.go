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
    ///
    topleft := image.Point{0, 0}
    bottomright := image.Point{2*width, 2*height}
    imgRGBA := image.NewRGBA(image.Rectangle{topleft, bottomright})
    ///

    r, g, b, a = [][]uint32{}, [][]uint32{}, [][]uint32{}, [][]uint32{}
    for y := 0; y < height; y++ {
        row_r, row_g, row_b, row_a := []uint32{}, []uint32{}, []uint32{}, []uint32{}

        for x := 0; x < width; x++ {
            nr, ng, nb, na := img.At(x, y).RGBA()
            ///
            col :=  color.RGBA{uint8(nr), uint8(ng), uint8(nb), uint8(na)}
            imgRGBA.Set(x, y, col)
            ///
            row_copies := 1;
            if x == 0 || x == width - 1{
                row_copies = 1     // append to font and back of row to allow interpolation on edge of image
            }
            for i := 0; i < row_copies; i++{
                row_r = append(row_r, nr)
                row_g = append(row_g, ng)
                row_b = append(row_b, nb)
                row_a = append(row_a, na)
                }
            }

        col_copies := 1
        if y == 0 || y == height - 1{
            col_copies = 1     // append to top and bottom of row to allow interpolation on edge of image
        }
        for i := 0; i < col_copies; i++{
            r = append(r, [][]uint32 {row_r}...)
            g = append(g, [][]uint32 {row_g}...)
            b = append(b, [][]uint32 {row_b}...)
            a = append(a, [][]uint32 {row_a}...)
        }
    }
    f, _ := os.Create("output_real.png")
    png.Encode(f, imgRGBA)
    return r, g, b, a, width, height, err
}

func cubicInterpolate(line []uint32)uint32{
    var mid uint32 = (line[1] + line[2])/2
    //var right uint32 = 3*(line[1] + line[2])/4
    //var left uint32 = (line[1] + line[2])/4
    /*
    newval := center*(line[2] - line[0] +center*(2.0*line[0] - 5.0*line[1] + 4.0*line[2] - line[3] + center*(3.0*(line[1] - line[2]) + line[3] - line[0])))
    newval = uint32(line[1] + newval/2)
    */
    //  f(x)/2 = ax^3 + bx^2 + cx +d 
    a := -line[0] + 3*line[1] - 3*line[2] + line[3]
    b := 2*line[0] - 5*line[1] + 4*line[2] - line[3]
    c := -line[0] + line[2]
    d := 2*line[1]
    f_mid :=uint32((a*(mid^3) + b*(mid^2) + c^mid + d)/2)
    //f_right :=uint32((a*(right^3) + b*(right^2) + c^right + d)/2) 
    //f_left :=uint32((a*(left^3) + b*(left^2) + c^left + d)/2) 
    ///new_points := []uint32{f_left, f_mid, f_right}
    return f_mid
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
    for y:= 2; y < height-2; y++ {
        for x := 2; x < width-2; x++{
            new_point := biCubicInterpolate(img[y-2:y+2][x-2:x+2])
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

    //curr_x := 2
    //curr_y := 2
    for y := 0; y < height; y++{
        for x := 0; x < width; x++{
            col :=  color.RGBA{uint8(r[y][x]), uint8(g[y][x]), uint8(b[y][x]), uint8(a[y][x])}
            imgRGBA.Set(2*x, 2*y, col)
        }
    }
    f, _ := os.Create("output.png")
    png.Encode(f, imgRGBA)
    /*
    for{
        pixelmap := <- ch
        p := pixelmap[MapKey{x:curr_x, y:curr_y}]
        if p.colors_added == 4 {
            //print the pixel on the screen
            col :=  color.RGBA{uint8(r[curr_y][curr_x]), uint8(g[curr_y][curr_x]), uint8(b[curr_y][curr_x]), uint8(a[curr_y][curr_x])}
            imgRGBA.Set(2*x, 2*y, col)

            //advance to next column or row
            if curr_x == width -2 -1{
                curr_x = 2
                curr_y++
            }
            if curr_y == height -2 {
                break
            }
        }

        ch <- pixelmap
    }
    */
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

    paintSection(ch, width, height, r, g, b, a)




}
