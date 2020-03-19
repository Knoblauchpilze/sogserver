FROM golang:1.13

WORKDIR /opt/dist

COPY sandbox /opt/dist

EXPOSE 3000

ENV LD_LIBRARY_PATH /opt/dist/lib

CMD /opt/dist/run.sh ${GO_ENV}
