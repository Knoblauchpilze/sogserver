FROM golang:1.13

WORKDIR /opt/dist

COPY sandbox /opt/dist

ENV LD_LIBRARY_PATH /opt/dist/lib
ENV APP_ENVIRONMENT development

EXPOSE 3000

CMD /opt/dist/run.sh ${APP_ENVIRONMENT}
