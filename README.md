# go_resize
  By Brian Neville
  
  This is an image resizing tool that I wrote.
  
  The function 'Resize' takes in the filename of a png image, and will create a version of 
  that file with the name 'resized_\<filename\>.png' and a resolution of twice the original.
  
  created on: go version go1.12.9 windows/amd64
  
  To use this, place the folder called 'resize' in the directory %GOROOT/src
  (e.g. the c:/go/src directory)
  
  import using:
  ```
  import(
    "resize"
    )
  ```
  
 resize images with the function:
 ```
  resize.Resize("filename.png")
 ```
 Note: this function currently only works on .png files
 
 <h1> Examples: </h1>
 

<h4>scaling flat colors</h4>

![img1](https://github.com/tropical-BN/go_resize/blob/master/sample_inputs_outputs/test_resizing.png)

![img11](https://github.com/tropical-BN/go_resize/blob/master/sample_inputs_outputs/resized_test_resizing.png)
 
 
<h4>scaling real images</h4>

![img21](https://github.com/tropical-BN/go_resize/blob/master/sample_inputs_outputs/reylo.png)


![img21](https://github.com/tropical-BN/go_resize/blob/master/sample_inputs_outputs/resized_reylo.png)

  
<h4>scaling png images with complex borders</h4>

![img31](https://github.com/tropical-BN/go_resize/blob/master/sample_inputs_outputs/flowers.png)


![img31](https://github.com/tropical-BN/go_resize/blob/master/sample_inputs_outputs/resized_flowers.png)
 
<h4>lenna (standard test image)</h4>
 
![img41](https://github.com/tropical-BN/go_resize/blob/master/sample_inputs_outputs/lenna/lenna.png)
 
![img41](https://github.com/tropical-BN/go_resize/blob/master/sample_inputs_outputs/lenna/resized_lenna.png)
 
 
<h4>progression gifs</h4>
  
![img0](https://github.com/tropical-BN/go_resize/blob/master/sample_inputs_outputs/lenna/partial.gif)
