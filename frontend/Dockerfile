# Stage-less build because the frontend is pure static assets.
FROM nginx:1.27-alpine

# Copy static HTMX assets into the default nginx web root.
COPY index.html /usr/share/nginx/html/index.html
COPY assets /usr/share/nginx/html/assets

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
