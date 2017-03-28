FROM alpine:3.5
MAINTAINER zweizeichen@element-43.com

RUN apk add --no-cache caddy yajl python3-dev g++ libffi-dev
RUN mkdir /top_stations
WORKDIR /top_stations

COPY ./requirements.txt /top_stations/requirements.txt
RUN pip3 install --upgrade pip && pip3 install -r requirements.txt

COPY . /top_stations

EXPOSE 8000
CMD ["/usr/bin/python3", "main.py"]