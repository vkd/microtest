# FROM scratch
FROM alpine

MAINTAINER Dmitry Kutakov <vkd.castle@gmail.com>

# EXPOSE 9000

ADD ./app /microtest/microtest

WORKDIR /microtest

ENTRYPOINT ["/microtest/microtest"]
CMD ["./microtests"]
