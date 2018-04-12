FROM 100.125.16.65:20202/hwcse/as-go:1.8.5

COPY ./test03 /home
COPY ./conf /home/conf
RUN chmod +x /home/test03

CMD ["/home/test03"]