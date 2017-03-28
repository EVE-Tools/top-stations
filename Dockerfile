FROM alpine:3.5
MAINTAINER zweizeichen@element-43.com

ENV PORT 80
EXPOSE 80

RUN apk add --no-cache caddy yajl python3-dev g++ libffi-dev
RUN mkdir /top-stations
WORKDIR /top-stations

COPY ./requirements.txt /top-stations/requirements.txt
RUN pip3 install --upgrade pip && pip3 install -r requirements.txt

COPY . /top-stations

CMD ["/usr/bin/python3", "main.py"]