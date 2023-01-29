FROM scratch
COPY markdown2confluence /markdown2confluence
ENTRYPOINT [ "/markdown2confluence" ]
