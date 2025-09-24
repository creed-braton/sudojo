FROM node:22.19.0 as build

ARG DOMAIN 
ENV VITE_DOMAIN=$DOMAIN
ARG DEVELOPMENT 
ENV VITE_DEV=$DEVELOPMENT

WORKDIR /app

COPY . .

RUN npm install && npm run build

FROM nginx:1.29.1

COPY --from=build /app/dist /usr/share/nginx/html
COPY ./nginx.conf ./etc/nginx/nginx.conf

CMD ["nginx", "-g", "daemon off;"]
