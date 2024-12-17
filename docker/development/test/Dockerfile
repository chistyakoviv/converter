FROM ubuntu:latest

RUN apt-get update 
RUN apt-get install -y \
    sudo \
    software-properties-common \
    build-essential \
    pkg-config \
    git \
    meson \
    ninja-build \
    wget 

RUN apt-get install -y \
    libglib2.0-dev \
    libexpat1-dev \
    librsvg2-dev \
    libpng-dev \
    libgif-dev \
    libjpeg-dev \
    libtiff-dev \
    libexif-dev \
    liblcms2-dev \
    liborc-dev \
    libffi-dev

RUN apt-get install -y \
    libaom-dev \
    ffmpeg

RUN apt-get install -y \
    cmake \
    libaom-dev \
    libdav1d-dev \
    libx265-dev \
    libyaml-cpp-dev \
    libssl-dev \
    libde265-dev \
    libx264-dev

# Need to be available at build time
ARG VIPS_VERSION=8.16.0
ARG VIPS_URL=https://github.com/libvips/libvips/releases/download
ARG HEIF_VERSION=1.19.5
ARG HEIF_URL=https://github.com/strukturag/libheif/releases/download

# Need to be available at run time and build time
ENV LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH

WORKDIR /usr/local/src

RUN wget ${HEIF_URL}/v${HEIF_VERSION}/libheif-${HEIF_VERSION}.tar.gz \
    && tar -xf libheif-${HEIF_VERSION}.tar.gz \
    && cd libheif-${HEIF_VERSION} \
    && mkdir build \
    && cd build \
    && cmake --preset=release .. \
    && make -j$(nproc) \
    && sudo make install

RUN wget ${VIPS_URL}/v${VIPS_VERSION}/vips-${VIPS_VERSION}.tar.xz \
    && tar -xf vips-${VIPS_VERSION}.tar.xz \
    && cd vips-${VIPS_VERSION} \
    && meson setup build --prefix=/usr \
    && meson compile -C build \
    && sudo meson install -C build
