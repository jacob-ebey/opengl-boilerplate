FROM golang:1.13

RUN apt update
RUN apt install -y libgl1-mesa-dev xorg-dev libc6-dev-i386
RUN apt install -y --reinstall build-essential