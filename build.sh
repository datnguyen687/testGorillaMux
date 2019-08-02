#!/bin/bash

echo "###### SET UP ######"
sudo docker build -t test_gorilla_mux .

echo "###### RUN ######"
sudo docker run -it -p 8080:8080 -v /home/dat/shared:/shared test_gorilla_mux:latest