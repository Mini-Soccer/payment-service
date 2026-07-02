# Stage 1: Build the application

# builder adalah host dimana aplikasi kita di-build
# ini adalah host, tempat untuk compile aplikasi, download dependency, build binary
# kenapa disebut bulder karena cuma dipakai untuk build binary, bukan untuk dijalankan di production
# stage 1 (stage builder) bertujuan untuk tempat compile aplikasi
# stage ini tidak dipakai di production karena:
# - image besar
# - ada compiler
# - ada source code
# - tidak aman kalau langsung dipakairun
FROM golang:1.25.5-alpine AS builder

# set environment variables
# dipakai untuk mengatur environment variable saat proses build go berjalan di dalam container (stage builder)
# GO111MODULE=on => memaksa go menggunakan go modules (dependency diambil dari go.mod) (tidak pakai GOPATH lama)
# CGO_ENABLED=0 => (penting) ini mematikan CGO, tidak tergantunGOARCH=amd64g library C di sistem
# GOOS=linux => menentukan target OS hasil build
# GOARCH=amd64 => menentukan arsitektur CPU target, arsitektur 64-bit x86 (umum di server & cloud)
# GODEBUG=http2client=0 => ini akan memaksa go tidak menggunakan HTTP/2 ke proxy.golang.org yang sering crash di container network
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GODEBUG=http2client=0

# set working directory
# kalau folder /app belum ada, docker akan otomatis membuatnya, mkdir -p /app; cd app
# semua perintah setelah ini (COPY, RUN, CMD, dll) akan dijalankan didalam folder /app didalam container
WORKDIR /app

# copy source code nya
# copy semua file dari folder project di host (saat ini) ke dalam container
# dari . (folder saat ini di host) -> ke . (working directory di container /app)
# nanti beberapa file tidak akan dicopy ke container, daftar file yg tidak dicopy ada di .dockerignore (env disini diignore juga)
#
COPY . .

# download dependencies
# jalankan perintah go mod download di dalam container saat proses build image
# perintah ini akan:
# - membaca go.mod
# - membaca go.sum
# - download semua dependency ke module cache (di docker)
# jadi dependency sudah tersedia sebelum go build
RUN go mod download

# build aplikasi
# compile aplikasi go
# hasilnya berupa file binary bernama payment-service
# dibuild dari source code di folder saatini (.)
RUN go build -o payment-service .

# Stage 2: Create production image
# adalah container final yang dijalankan di server/cloud
# prinsipnya ambil hasil build (binary saja) -> jalankan di environment minimal, aman, dan ringan
# fungsinya hanya menjalankan binary ./payment-service yang sudah dibuat di stage 1
# - tidak ada go compiler
# - tidak ada source code
# - pakai host alpine jadi lebih kecil
# - aplikasi berjalan sebagai user biasa (bukan root), jadi lebih aman
# - di production kita tidak butuh compiler go
# - Install wkhtmltopdf di final image (Debian-based)
FROM debian:bookworm-slim

# install wkhtmltopdf + full dependencies (FIX rendering issue)
RUN apt-get update && apt-get install -y \
    wkhtmltopdf \
    ca-certificates \
    fonts-dejavu \
    fonts-liberation \
    fontconfig \
    libxrender1 \
    libxext6 \
    libx11-6 \
    libjpeg62-turbo \
    libpng16-16 \
    xfonts-base \
    xfonts-75dpi \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

# Set the timezone environment variable (can be overridden by .env)
ENV TZ=Asia/Jakarta

# config timezone
RUN ln -sf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# set working directory
# semua operasi selanjutnya terjadi di /app
# binary akan berada di /app/payment-service
WORKDIR /app

# create user and group for application
# membuat group dengan GID 1001
# membuat user:
# - UID:1001
# - group:binarygroup
# - tanpa passowrd (-D)
#
# secara default container berjalan sebagai root (DEBIAN)
# kalau terjadi explpit:
# - hacker dapat akses root
# - bisa modifikasi filesystem container
# - lebih berbahaya
# dengan USER userapp, aplikasi berjalan sebagai user biasa (best practice security di production)
RUN groupadd -g 1001 binarygroup && \
    useradd -u 1001 -g binarygroup -m userapp

# XDG runtime
ENV XDG_RUNTIME_DIR=/tmp/runtime-userapp
RUN install -d -m 700 -o userapp -g binarygroup /tmp/runtime-userapp

# BUAT DIR + SET OWNER (INI YANG FIX)
RUN mkdir -p /tmp/runtime-userapp && \
    chown userapp:binarygroup /tmp/runtime-userapp && \
    chmod 700 /tmp/runtime-userapp

# copy the binary from the builder stage
# ini bagian inti multi-stage:
# - artinya ambil file /app/payment-service dari stage 1 (builder)
# - copy ke stage ini (stage 2)
# - sekaligus ubah owner ke userapp:binarygroup
# - pakai --chown supaya file tidak dimiliki root, user userapp bisa execute binary
COPY --from=builder --chown=userapp:binarygroup /app/payment-service .

# template html invoice (diluar binary)
COPY --from=builder --chown=userapp:binarygroup /app/template ./template

# switch to the userapp user
# semua perintah setelah ini dijalankan sebagai user userapp, bukan root, lebih aman
USER userapp

# expose port 8003
# ini hanya dokumentasi bahwa container listen di port 8003
EXPOSE 8003

# command to run the application
# ini perintah default saat container dijalankan
CMD ["./payment-service"]
