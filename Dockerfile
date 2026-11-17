FROM alpine:3.24
ENV LANG=C.UTF-8 TZ=Asia/Shanghai
RUN apk --no-cache add tzdata
WORKDIR /app
COPY dist/qingdao_backend ./
EXPOSE 3010
ENTRYPOINT ["./qingdao_backend"]
