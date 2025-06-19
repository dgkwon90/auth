# 1단계: 빌드 스테이지
FROM --platform=$BUILDPLATFORM golang:1.24.2 AS builder

# 환경 변수 설정
ENV CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /app

# go.mod, go.sum 복사 및 의존성 설치
COPY go.mod go.sum ./
RUN go mod download

# 소스 코드 복사
COPY . .

# main.go 빌드 (cmd/main.go 기준)
RUN GOARCH=$(go env GOARCH) GOOS=$GOOS go build -o auth ./cmd/main.go

# 2단계: 런타임 스테이지 (distroless 이미지 사용)
FROM alpine:3.20

WORKDIR /app

# 빌드된 바이너리 복사
COPY --from=builder /app/auth .

# 컨테이너 시작 시 실행될 명령어
ENTRYPOINT ["/app/auth"]
