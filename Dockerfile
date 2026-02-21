FROM node:22.19.0 AS frontend

WORKDIR /build

COPY frontend/package.json frontend/package-lock.json ./

RUN npm install

COPY frontend .

RUN npm run build

FROM golang:1.25.1-alpine3.22 AS backend

WORKDIR /build

COPY backend/go.mod backend/go.sum ./

RUN go mod download

COPY backend/main.go ./main.go
COPY backend/pkg ./pkg
COPY backend/adp ./adp
COPY backend/svc ./svc

RUN go build -o ./bin ./main.go

FROM alpine:3.22

COPY --from=backend /build/bin /main
COPY --from=frontend /build/dist /static

ENV PORT=8080

ENTRYPOINT [ "/main" ]
