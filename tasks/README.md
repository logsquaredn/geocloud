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
```sh
# reproject:
../bin/reproject /home/phish3y/Downloads/input_shape/zip/input_shape.zip /home/phish3y/Downloads/output_shape 2000
# buffer:
../bin/buffer /home/phish3y/Downloads/input_shape/zip/input_shape.zip /home/phish3y/Downloads/output_shape 2 50
# filter:
../bin/filter /home/phish3y/Downloads/input_shape/zip/input_shape.zip /home/phish3y/Downloads/output_shape 'BASIN' 'al'
# remove bad geometry:
../bin/removebadgeometry /home/phish3y/Downloads/input_shape/zip/input_shape.zip /home/phish3y/Downloads/output_shape
```
