// Entrypoint for the kitty service. For the Phase 0 skeleton it exposes only
// health endpoints; the Domain / Application / Infrastructure layers and the
// expense/split/settlement logic arrive from Phase 3 onwards.

var builder = WebApplication.CreateBuilder(args);

var app = builder.Build();

// Routes live under /api/v1/kitty so Traefik can path-route without rewriting.
var kitty = app.MapGroup("/api/v1/kitty");
kitty.MapGet("/healthz", () => Results.Ok(new HealthResponse("ok", "kitty")));
kitty.MapGet("/readyz", () => Results.Ok(new HealthResponse("ok", "kitty")));

app.Run();

internal record HealthResponse(string Status, string Service);

// Exposed so the integration test project can drive the app via
// WebApplicationFactory<Program>.
public partial class Program { }
