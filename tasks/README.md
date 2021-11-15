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

### TODO
- Fix bad geometries
- dissolve
- point in poly and poly in poly lookups

### Examples
```sh
# reproject:
./bin/reproject /home/evan/Downloads/input_shape/AL112017_windswath.shp /home/evan/Downloads/output_shape/AL112017_windswath_reprojected.shp 2000
# buffer:
./bin/buffer /home/evan/Downloads/input_shape/AL112017_windswath.shp /home/evan/Downloads/output_shape/AL112017_windswath_buffered.shp 2
# filter:
./bin/filter /home/evan/Downloads/input_shape/AL112017_windswath.shp /home/evan/Downloads/output_shape/AL112017_windswath_filtered.shp 'BASIN' 'al'
# remove bad geometry:
./bin/removebadgeometry /home/evan/Downloads/input_shape/AL112017_windswath.shp /home/evan/Downloads/output_shape/AL112017_windswath_goodGeometry.shp
```
