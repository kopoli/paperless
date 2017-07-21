FROM ubuntu:16.04

RUN \
    /bin/sed -i -e 's,http://archive.ubuntu.com/ubuntu,mirror://mirrors.ubuntu.com/mirrors.txt,g' /etc/apt/sources.list && \
    apt-get update && \
    apt-get -y upgrade && \
    apt-get install -y unpaper imagemagick tesseract-ocr tesseract-ocr-fin && \
    rm -rf /var/lib/apt/lists/*

EXPOSE 8078

CMD [ "/paperless" ]
