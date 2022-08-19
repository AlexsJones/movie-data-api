#!/bin/bash
for i in 00*; do
    aws s3 cp $i s3://movie-data-api/$i
done

