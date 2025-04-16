FROM grafana/grafana:latest

# Copiar archivos de configuracion
COPY ./dashboard.json /var/lib/grafana/dashboards/dashboard.json
COPY ./dashboard.yaml /etc/grafana/provisioning/dashboards/dashboard.yaml
COPY ./datasource.yaml /etc/grafana/provisioning/datasources/datasource.yaml

# instalar el plugin de grafana
# RUN grafana-cli plugins install redis-datasource
ENV GF_INSTALL_PLUGINS=redis-datasource
# ENV GF_SECURITY_ADMIN_USER=admin  
# ENV GF_SECURITY_ADMIN_PASSWORD=sopes1  

# Exponer el puerto 3000
EXPOSE 3000

