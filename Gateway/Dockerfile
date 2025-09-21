FROM mcr.microsoft.com/dotnet/sdk:9.0 AS build
WORKDIR /src

COPY ["Gateway.csproj", "."]

RUN dotnet restore "Gateway.csproj"
COPY . .
RUN dotnet build "Gateway.csproj" -c Release -o /app/build

FROM mcr.microsoft.com/dotnet/aspnet:9.0 AS runtime
WORKDIR /app
EXPOSE 8080

COPY --from=build /app/build .

ENTRYPOINT ["dotnet", "Gateway.dll"]
