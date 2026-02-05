# syntax=docker/dockerfile:1.4

FROM golang:1.23-alpine AS builder

# git 설치 (버전 정보용)
RUN apk add --no-cache git

WORKDIR /app

# 의존성 캐싱 레이어
COPY go.mod go.sum ./

# go mod 캐시 마운트로 재사용
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# 소스 코드 복사
COPY . .

# 빌드 캐시 마운트로 컴파일 속도 향상
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" \
    -trimpath \
    -o main ./cmd/server

# 최종 이미지
FROM scratch

WORKDIR /app

# CA 인증서 복사
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 바이너리와 정적 파일만 복사
COPY --from=builder /app/main .
COPY --from=builder /app/web ./web

# 헬스체크용 포트 노출
EXPOSE 8080

CMD ["/app/main"]
