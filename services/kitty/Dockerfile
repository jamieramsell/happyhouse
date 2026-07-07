# Publish the ASP.NET Core app, then run it on the smaller runtime image.
FROM mcr.microsoft.com/dotnet/sdk:10.0 AS build
WORKDIR /src
COPY src/Kitty.Api/Kitty.Api.csproj src/Kitty.Api/
RUN dotnet restore src/Kitty.Api/Kitty.Api.csproj
COPY src/ src/
RUN dotnet publish src/Kitty.Api/Kitty.Api.csproj -c Release -o /out

FROM mcr.microsoft.com/dotnet/aspnet:10.0 AS runtime
WORKDIR /app
COPY --from=build /out ./
# Kestrel listens on 8080 to match the other services behind Traefik.
ENV ASPNETCORE_URLS=http://+:8080
EXPOSE 8080
ENTRYPOINT ["dotnet", "Kitty.Api.dll"]
