FROM alpine:3.6

ENV PORT 8000
EXPOSE 8000

RUN apk add --no-cache caddy yajl python3-dev g++ libffi-dev
RUN mkdir /top-stations
WORKDIR /top-stations

COPY ./requirements.txt /top-stations/requirements.txt
RUN pip3 install --upgrade pip && pip3 install -r requirements.txt

# Do not run as root
USER element43:element43

COPY . /top-stations

CMD ["/usr/bin/python3", "main.py"]