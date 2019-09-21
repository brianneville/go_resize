import imageio
import os

def generate_fnames(original_width, original_height):
    assert original_width > 2 and original_height > 2
    y = 0
    x = 0
    for y in range(0, original_height-1):
        for x in range(0, original_width -1):
            x_p = str(x).zfill(3)		# format according to filename convention that i had used e.g. r_0010291.png
            y_p = str(y).zfill(3)
            for i in range(0, 3):
                f_name = f"r_{y_p}{x_p}{i}.png"
                yield f_name


Gen = generate_fnames(50, 50)		# size of input image is 50x50 pixels
with imageio.get_writer(os.getcwd()+'/progression/video_small_test_2.gif', mode='I') as writer:
    for filename in Gen:
        print(filename)
        img = imageio.imread(filename)
        writer.append_data(img)