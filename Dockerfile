# ======================================
# Main dependencies
# ======================================
ARG GO_VERSION=1.25
ARG FFMPEG_VERSION=8.0
ARG YTDLP_VERSION=2025.08.22

ARG BUILD_VERSION=development

# ======================================
# Build Go application
# ======================================
FROM golang:${GO_VERSION}-bookworm AS build_app
ARG BUILD_VERSION
WORKDIR /app
RUN apt-get install -y --no-install-recommends git
COPY . /app
ENV GO111MODULE=on
ENV CGO_ENABLED=0
RUN go build -tags urfave_cli_no_docs -ldflags "-X github.com/exler/yt-transcribe/cmd.Version=${BUILD_VERSION}" -o /yt-transcribe

# ======================================
# Build whisper.cpp and FFmpeg (with whisper)
# ======================================
FROM debian:bookworm-slim AS build_ffmpeg
ARG FFMPEG_VERSION
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    pkg-config \
    git \
    ca-certificates \
    curl \
    cmake \
    yasm \
    nasm \
    libtool \
    autoconf \
    automake \
    xz-utils \
    bzip2 \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /tmp

# 1) Build & install whisper.cpp to /usr/local (shared + static)
RUN git clone --depth 1 https://github.com/ggml-org/whisper.cpp.git whisper.cpp && \
    cmake -S whisper.cpp -B whisper.cpp/build \
      -DCMAKE_BUILD_TYPE=Release \
      -DWHISPER_BUILD_EXAMPLES=OFF \
      -DWHISPER_BUILD_TESTS=OFF \
      -DWHISPER_BUILD_SHARED_LIB=ON \
      -DWHISPER_BUILD_STATIC_LIB=ON \
      -DCMAKE_INSTALL_PREFIX=/usr/local && \
    cmake --build whisper.cpp/build -j"$(nproc)" && \
    cmake --install whisper.cpp/build

# 2) Build & install FFmpeg with --enable-whisper (no external codec deps)
RUN curl -sSL "https://ffmpeg.org/releases/ffmpeg-${FFMPEG_VERSION}.tar.xz" -o ffmpeg.tar.xz; \
    tar -xf ffmpeg.tar.xz; \
    cd "ffmpeg-${FFMPEG_VERSION}"; \
    PKG_CONFIG_PATH=/usr/local/lib/pkgconfig ./configure \
      --prefix=/usr/local \
      --disable-debug \
      --disable-doc \
      --disable-ffplay \
      --enable-whisper; \
    make -j"$(nproc)"; \
    make install; \
    ldconfig

# ======================================
# Fetch yt-dlp prebuilt binary
# ======================================
FROM debian:bookworm-slim AS fetch_ytdlp
ARG YTDLP_VERSION
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates wget && rm -rf /var/lib/apt/lists/*
WORKDIR /build
RUN wget -q -O yt-dlp "https://github.com/yt-dlp/yt-dlp/releases/download/${YTDLP_VERSION}/yt-dlp_linux" && \
    chmod +x yt-dlp

# ======================================
# Final runtime image
# ======================================
FROM debian:bookworm-slim
ENV DEBIAN_FRONTEND=noninteractive
WORKDIR /app

# Runtime deps (keep minimal)
RUN apt-get update && apt-get install -y --no-install-recommends \
    wget \
    libstdc++6 \
    zlib1g \
    libgomp1 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && update-ca-certificates

# Whisper model
RUN mkdir -p /app/models && wget https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.bin -O /app/models/ggml-small.bin

ENV LD_LIBRARY_PATH=/usr/local/lib
ENV PATH="/usr/local/bin:/usr/bin:/bin"

# App
COPY --from=build_app /yt-transcribe /app/yt-transcribe

# FFmpeg + libs
COPY --from=build_ffmpeg /usr/local/bin/ffmpeg  /usr/local/bin/ffmpeg
COPY --from=build_ffmpeg /usr/local/bin/ffprobe /usr/local/bin/ffprobe
COPY --from=build_ffmpeg /usr/local/lib/       /usr/local/lib/

# yt-dlp
COPY --from=fetch_ytdlp /build/yt-dlp /usr/local/bin/yt-dlp

ENTRYPOINT ["/app/yt-transcribe"]
CMD ["runserver"]

EXPOSE 8000
