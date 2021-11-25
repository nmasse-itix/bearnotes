FROM scratch
ARG BUILT_ARTIFACT
ADD "$BUILT_ARTIFACT" /
ENTRYPOINT [ "/bearnotes-migrator" ]
CMD [ ]

