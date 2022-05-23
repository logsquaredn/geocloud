# task

## Developing

### Prerequisites

gdal is *required* - install instructions:

```sh
sudo add-apt-repository ppa:ubuntugis/ppa
sudo apt-get update
sudo apt-get install gdal-bin
sudo apt-get install libgdal-dev
export C_INCLUDE_PATH=/usr/include/gdal
```
### Examples
Local build:
```sh
gcc -Wall tasks/buffer/buffer.c tasks/shared/shared.c -l gdal -o assets/buffer
gcc -Wall tasks/filter/filter.c tasks/shared/shared.c -l gdal -o assets/filter
gcc -Wall tasks/reproject/reproject.c tasks/shared/shared.c -l gdal -o assets/reproject
gcc -Wall tasks/removebadgeometry/removebadgeometry.c tasks/shared/shared.c -l gdal -o assets/removebadgeometry
gcc -Wall tasks/lookup/vectorlookup.c tasks/shared/shared.c -l gdal -o assets/vectorlookup
gcc -Wall tasks/lookup/rasterlookup.c tasks/shared/shared.c -l gdal -o assets/rasterlookup
```

Local run:
```sh
../assets/reproject /home/phish3y/Documents/input_shape/zip/input_shape.zip /home/phish3y/Documents/output_shape 2000
../assets/buffer /home/phish3y/Documents/input_shape/zip/input_shape.zip /home/phish3y/Documents/output_shape 2 50
../assets/filter /home/phish3y/Documents/input_shape/zip/input_shape.zip /home/phish3y/Documents/output_shape 'BASIN' 'al'
../assets/removebadgeometry /home/phish3y/Documents/input_shape/zip/input_shape.zip /home/phish3y/Documents/output_shape
../assets/vectorlookup /home/phish3y/Documents/input_shape/mmi/mmi.zip /home/phish3y/Documents/output_shape 97.5679 34.6970
../assets/rasterlookup /home/phish3y/Documents/input_shape/input.tif /home/phish3y/Documents/output_shape 2,4 -89.9383 29.95313
```
