FROM hashicorp/vault

COPY ./init.sh /

RUN chmod +x init.sh

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD [ "/init.sh" ]