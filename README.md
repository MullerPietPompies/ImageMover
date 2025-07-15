# File Mover Utility

## Purpose 
I was approached by my father with a proplem his company faces when working with images and files. For context the company uses airial drone photography and AI to count game and other animals. As the model processes the the images to identify game the location data caputred by the drone gets lost. This small utility uses the file structure produced by the Computer Vision Model and duplicates that same file structure but replacing each image produced with the corresponding original image from the drone. The original images with the location data is used to plot it on the map of the particular farm or park thats being counted so the client can see where each of the animals were located on the property

## How it works
- It takes in a `blueprint` file path.
  - This contains the sorted output of the AI model in folders such as (this is a small exmaple of what the program expects.)
 ```
├── giraffe
│   ├── DJI_001.png
│   ├── DJI_022.png
│   ├── DJI_453.png
│   └── DJI_654.png
└── rhino
    ├── DJI_098.png
    └── DJI_205.png
```
- A `image directory` is then selected
  - This is the folder containing the raw images directly from the SD card

- A `destination` is selected this is where the folder structure will be duplacted to

## Installation Guide

Simply download the build for your system on the releases page and run it. No instalation process is required

## Development

- In the source code is a provided `flake.nix` file that fill install all the dependencies needed to work on this project
-  All code is in the `src` directory.
