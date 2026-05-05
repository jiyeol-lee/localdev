FROM alpine:3.20

ENV PANE_NAME=localdev-pane

CMD ["sh", "-c", "i=1; while true; do printf '%s tick %04d\\n' \"$PANE_NAME\" \"$i\"; i=$((i + 1)); sleep 1; done"]
