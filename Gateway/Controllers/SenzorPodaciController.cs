using System;
using System.Linq;
using System.Threading;
using System.Collections.Generic;
using Microsoft.AspNetCore.Mvc;
using Gateway.Models;
using Gateway.GrpcClient;
using Grpc.Net.Client;
using Google.Protobuf.WellKnownTypes;

namespace Gateway.Controllers;

[ApiController]
[Route("api/[Controller]")]
public class SenzorPodaciController : ControllerBase
{
	private SenzorPodaci.SenzorPodaciClient client;
	private CancellationToken cancellationToken;

	public SenzorPodaciController()
	{
		var chan = GrpcChannel.ForAddress("http://datamanager:5050", new GrpcChannelOptions
		{
			HttpHandler = new HttpClientHandler(),
		});
		this.client = new SenzorPodaci.SenzorPodaciClient(chan);
		this.cancellationToken = new CancellationTokenSource(TimeSpan.FromSeconds(30)).Token;
	}

	[HttpGet("VratiSenzorPodatak/{id:int}")]
	public async Task<IActionResult>VratiSenzorPodatak(int id)
	{
		Console.WriteLine($"Pribavljanje podatka sa ID: {id}");
		try
		{
			Int32Value ID = new Int32Value { Value = id };
			var rezultat = await client.VratiSenzorPodatakAsync(ID, cancellationToken: cancellationToken);

			Console.WriteLine($"Podatak sa ID: {id} pronadjen.");
			return Ok(rezultat);
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Greška: {ex.Message}");
			if (ex.Message.Length > 6 && ex.Message.Substring(0, 6) == "Status")
			{
				return ParseStatus(ex.Message);
			}
			return BadRequest(ex.Message);
		}
	}

	[HttpGet("SviSenzorPodaci")]
	public async Task<IActionResult>SviSenzorPodaci()
	{
		Console.WriteLine("Pribavljanje svih podataka.");
		try
		{
			List<Gateway.GrpcClient.SenzorPodatak> rezultat = new();
			using var call = client.SviSenzorPodaci(new Google.Protobuf.WellKnownTypes.Empty());

			while (await call.ResponseStream.MoveNext(cancellationToken: cancellationToken))
			{
				rezultat.Add(call.ResponseStream.Current);
			}

			Console.WriteLine("Svi podaci uspešno pribavljeni.");
			return Ok(rezultat);
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Greška: {ex.Message}");
			if (ex.Message.Length > 6 && ex.Message.Substring(0, 6) == "Status")
			{
				return ParseStatus(ex.Message);
			}
			return BadRequest(ex.Message);
		}
	}

	[HttpPost("DodajSenzorPodatak")]
	public async Task<IActionResult>DodajSenzorPodatak([FromBody] Gateway.GrpcClient.SenzorPodatak podatak)
	{
		Console.WriteLine("Dodavanje novog podatka.");
		try
		{
			await client.DodajSenzorPodatakAsync(podatak, cancellationToken: cancellationToken);
			Console.WriteLine("Podatak uspešno dodat.");
			return Ok("Podatak uspešno dodat.");
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Greška: {ex.Message}");
			if (ex.Message.Length > 6 && ex.Message.Substring(0, 6) == "Status")
			{
				return ParseStatus(ex.Message);
			}
			return BadRequest(ex.Message);
		}
	}

	[HttpPut("IzmeniSenzorPodatak")]
	public async Task<IActionResult>IzmeniSenzorPodatak([FromBody] Gateway.GrpcClient.SenzorPodatak podatak)
	{
		Console.WriteLine($"Izmena podatka sa ID: {podatak.Id}");
		try
		{
			await client.IzmeniSenzorPodatakAsync(podatak, cancellationToken: cancellationToken);
			Console.WriteLine($"Podatak sa ID {podatak.Id} uspešno izmenjen.");
			return Ok($"Podatak sa id {podatak.Id} uspešno izmenjen.");
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Greška: {ex.Message}");
			if (ex.Message.Length > 6 && ex.Message.Substring(0, 6) == "Status")
			{
				return ParseStatus(ex.Message);
			}
			return BadRequest(ex.Message);
		}
	}

	[HttpDelete("IzbrisiSenzorPodatak/{id:int}")]
	public async Task<IActionResult>IzbrisiSenzorPodatak(int id)
	{
		Console.WriteLine($"Brisanje podatka sa ID: {id}");
		try
		{
			Int32Value ID = new Int32Value { Value = id };
			await client.IzbrisiSenzorPodatakAsync(ID, cancellationToken: cancellationToken);
			Console.WriteLine($"Podatak sa ID: {id} uspešno izbrisan.");
			return Ok($"Podatak sa ID: {id} uspešno izbrisan.");
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Greška: {ex.Message}");
			if (ex.Message.Length > 6 && ex.Message.Substring(0, 6) == "Status")
			{
				return ParseStatus(ex.Message);
			}
			return BadRequest(ex.Message);
		}
	}

/////////////////////////////////////////////////////////////////////////////////////

	[HttpPost("MinTemperatura")]
	public async Task<IActionResult>MinTemperatura([FromBody]Gateway.GrpcClient.VremenskiPeriod period)
	{
		Console.WriteLine($"Pribavljanje minimalne temerature u periodu izmedju {period.Pocetak} i {period.Kraj}");
		try
		{
			List<Gateway.GrpcClient.SenzorPodatak> rezultat = new();
			using var call = client.SviSenzorPodaciPeriod(period);

			while (await call.ResponseStream.MoveNext(cancellationToken: cancellationToken))
			{
				rezultat.Add(call.ResponseStream.Current);
			}

			Console.WriteLine($"Uspešno pribavljena minimalna temperatura za period izmedju {period.Pocetak} i {period.Kraj}.");
			return Ok(rezultat.Min(p => p.Temperatura));
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Greška: {ex.Message}");
			if (ex.Message.Length > 6 && ex.Message.Substring(0, 6) == "Status")
			{
				return ParseStatus(ex.Message);
			}
			return BadRequest(ex.Message);
		}
	}

	[HttpPost("MaxTemperatura")]
	public async Task<IActionResult>MaxTemperatura([FromBody]Gateway.GrpcClient.VremenskiPeriod period)
	{
		Console.WriteLine($"Pribavljanje maksimalne temerature u periodu izmedju {period.Pocetak} i {period.Kraj}");
		try
		{
			List<Gateway.GrpcClient.SenzorPodatak> rezultat = new();
			using var call = client.SviSenzorPodaciPeriod(period);

			while (await call.ResponseStream.MoveNext(cancellationToken: cancellationToken))
			{
				rezultat.Add(call.ResponseStream.Current);
			}

			Console.WriteLine($"Uspešno pribavljena maksimalna temperatura za period izmedju {period.Pocetak} i {period.Kraj}.");
			return Ok(rezultat.Max(p => p.Temperatura));
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Greška: {ex.Message}");
			if (ex.Message.Length > 6 && ex.Message.Substring(0, 6) == "Status")
			{
				return ParseStatus(ex.Message);
			}
			return BadRequest(ex.Message);
		}
	}

	[HttpPost("AvgTemperatura")]
	public async Task<IActionResult>AvgTemperatura([FromBody]Gateway.GrpcClient.VremenskiPeriod period)
	{
		Console.WriteLine($"Pribavljanje prosečne temerature u periodu izmedju {period.Pocetak} i {period.Kraj}");
		try
		{
			List<Gateway.GrpcClient.SenzorPodatak> rezultat = new();
			using var call = client.SviSenzorPodaciPeriod(period);

			while (await call.ResponseStream.MoveNext(cancellationToken: cancellationToken))
			{
				rezultat.Add(call.ResponseStream.Current);
			}

			Console.WriteLine($"Uspešno pribavljena prosečna temperatura za period izmedju {period.Pocetak} i {period.Kraj}.");
			return Ok(rezultat.Average(p => p.Temperatura));
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Greška: {ex.Message}");
			if (ex.Message.Length > 6 && ex.Message.Substring(0, 6) == "Status")
			{
				return ParseStatus(ex.Message);
			}
			return BadRequest(ex.Message);
		}
	}

	[HttpPost("SumTemperatura")]
	public async Task<IActionResult>SumTemperatura([FromBody]Gateway.GrpcClient.VremenskiPeriod period)
	{
		Console.WriteLine($"Pribavljanje sume temerature u periodu izmedju {period.Pocetak} i {period.Kraj}");
		try
		{
			List<Gateway.GrpcClient.SenzorPodatak> rezultat = new();
			using var call = client.SviSenzorPodaciPeriod(period);

			while (await call.ResponseStream.MoveNext(cancellationToken: cancellationToken))
			{
				rezultat.Add(call.ResponseStream.Current);
			}

			Console.WriteLine($"Uspešno pribavljena suma temperatura za period izmedju {period.Pocetak} i {period.Kraj}.");
			return Ok(rezultat.Sum(p => p.Temperatura));
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Greška: {ex.Message}");
			if (ex.Message.Length > 6 && ex.Message.Substring(0, 6) == "Status")
			{
				return ParseStatus(ex.Message);
			}
			return BadRequest(ex.Message);
		}
	}

/////////////////////////////////////////////////

	[HttpPost("MinVlaznostVazduha")]
	public async Task<IActionResult>MinVlaznostVazduha([FromBody]Gateway.GrpcClient.VremenskiPeriod period)
	{
		Console.WriteLine($"Pribavljanje minimalnog procenta vlažnosti vazduha u periodu izmedju {period.Pocetak} i {period.Kraj}");
		try
		{
			List<Gateway.GrpcClient.SenzorPodatak> rezultat = new();
			using var call = client.SviSenzorPodaciPeriod(period);

			while (await call.ResponseStream.MoveNext(cancellationToken: cancellationToken))
			{
				rezultat.Add(call.ResponseStream.Current);
			}

			Console.WriteLine($"Uspešno pribavljen minimalan procenat vlažnosti vazduha za period izmedju {period.Pocetak} i {period.Kraj}.");
			return Ok(rezultat.Min(p => p.VlaznostVazduha));
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Greška: {ex.Message}");
			if (ex.Message.Length > 6 && ex.Message.Substring(0, 6) == "Status")
			{
				return ParseStatus(ex.Message);
			}
			return BadRequest(ex.Message);
		}
	}

	[HttpPost("MaxVlaznostVazduha")]
	public async Task<IActionResult>MaxVlaznostVazduha([FromBody]Gateway.GrpcClient.VremenskiPeriod period)
	{
		Console.WriteLine($"Pribavljanje maksimalog procenta vlažnosti vazduha u periodu izmedju {period.Pocetak} i {period.Kraj}");
		try
		{
			List<Gateway.GrpcClient.SenzorPodatak> rezultat = new();
			using var call = client.SviSenzorPodaciPeriod(period);

			while (await call.ResponseStream.MoveNext(cancellationToken: cancellationToken))
			{
				rezultat.Add(call.ResponseStream.Current);
			}

			Console.WriteLine($"Uspešno pribavljen maksimalan procenat vlažnosti vazduha za period izmedju {period.Pocetak} i {period.Kraj}.");
			return Ok(rezultat.Max(p => p.VlaznostVazduha));
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Greška: {ex.Message}");
			if (ex.Message.Length > 6 && ex.Message.Substring(0, 6) == "Status")
			{
				return ParseStatus(ex.Message);
			}
			return BadRequest(ex.Message);
		}
	}

	[HttpPost("AvgVlaznostVazduha")]
	public async Task<IActionResult>AvgVlaznostVazduha([FromBody]Gateway.GrpcClient.VremenskiPeriod period)
	{
		Console.WriteLine($"Pribavljanje prosečnog procenta vlažnosti vazduha u periodu izmedju {period.Pocetak} i {period.Kraj}");
		try
		{
			List<Gateway.GrpcClient.SenzorPodatak> rezultat = new();
			using var call = client.SviSenzorPodaciPeriod(period);

			while (await call.ResponseStream.MoveNext(cancellationToken: cancellationToken))
			{
				rezultat.Add(call.ResponseStream.Current);
			}

			Console.WriteLine($"Uspešno pribavljen prosečan procenat vlažnosti vazduha za period izmedju {period.Pocetak} i {period.Kraj}.");
			return Ok(rezultat.Average(p => p.VlaznostVazduha));
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Greška: {ex.Message}");
			if (ex.Message.Length > 6 && ex.Message.Substring(0, 6) == "Status")
			{
				return ParseStatus(ex.Message);
			}
			return BadRequest(ex.Message);
		}
	}

	[HttpPost("SumVlaznostVazduha")]
	public async Task<IActionResult>SumVlaznostVazduha([FromBody]Gateway.GrpcClient.VremenskiPeriod period)
	{
		Console.WriteLine($"Pribavljanje sume procenta vlažnosti vazduha u periodu izmedju {period.Pocetak} i {period.Kraj}");
		try
		{
			List<Gateway.GrpcClient.SenzorPodatak> rezultat = new();
			using var call = client.SviSenzorPodaciPeriod(period);

			while (await call.ResponseStream.MoveNext(cancellationToken: cancellationToken))
			{
				rezultat.Add(call.ResponseStream.Current);
			}

			Console.WriteLine($"Uspešno pribavljena suma procenta vlažnosti vazduha za period izmedju {period.Pocetak} i {period.Kraj}.");
			return Ok(rezultat.Sum(p => p.VlaznostVazduha));
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Greška: {ex.Message}");
			if (ex.Message.Length > 6 && ex.Message.Substring(0, 6) == "Status")
			{
				return ParseStatus(ex.Message);
			}
			return BadRequest(ex.Message);
		}
	}

//////////////////////////////////////////////////////////

	[HttpPost("MinPm2_5")]
	public async Task<IActionResult>MinPm2_5([FromBody]Gateway.GrpcClient.VremenskiPeriod period)
	{
		Console.WriteLine($"Pribavljanje minimalne količine pm2.5 čestica u periodu izmedju {period.Pocetak} i {period.Kraj}");
		try
		{
			List<Gateway.GrpcClient.SenzorPodatak> rezultat = new();
			using var call = client.SviSenzorPodaciPeriod(period);

			while (await call.ResponseStream.MoveNext(cancellationToken: cancellationToken))
			{
				rezultat.Add(call.ResponseStream.Current);
			}

			Console.WriteLine($"Uspešno pribavljena minimalana količina pm2.5 čestica za period izmedju {period.Pocetak} i {period.Kraj}.");
			return Ok(rezultat.Min(p => p.Pm25));
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Greška: {ex.Message}");
			if (ex.Message.Length > 6 && ex.Message.Substring(0, 6) == "Status")
			{
				return ParseStatus(ex.Message);
			}
			return BadRequest(ex.Message);
		}
	}

	[HttpPost("MaxPm2_5")]
	public async Task<IActionResult>MaxPm2_5([FromBody]Gateway.GrpcClient.VremenskiPeriod period)
	{
		Console.WriteLine($"Pribavljanje maksimalne količine pm2.5 čestica u periodu izmedju {period.Pocetak} i {period.Kraj}");
		try
		{
			List<Gateway.GrpcClient.SenzorPodatak> rezultat = new();
			using var call = client.SviSenzorPodaciPeriod(period);

			while (await call.ResponseStream.MoveNext(cancellationToken: cancellationToken))
			{
				rezultat.Add(call.ResponseStream.Current);
			}

			Console.WriteLine($"Uspešno pribavljena količine pm2.5 čestica za period izmedju {period.Pocetak} i {period.Kraj}.");
			return Ok(rezultat.Max(p => p.Pm25));
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Greška: {ex.Message}");
			if (ex.Message.Length > 6 && ex.Message.Substring(0, 6) == "Status")
			{
				return ParseStatus(ex.Message);
			}
			return BadRequest(ex.Message);
		}
	}

	[HttpPost("AvgPm2_5")]
	public async Task<IActionResult>AvgPm2_5([FromBody]Gateway.GrpcClient.VremenskiPeriod period)
	{
		Console.WriteLine($"Pribavljanje prosečne količine pm2.5 čestica u periodu izmedju {period.Pocetak} i {period.Kraj}");
		try
		{
			List<Gateway.GrpcClient.SenzorPodatak> rezultat = new();
			using var call = client.SviSenzorPodaciPeriod(period);

			while (await call.ResponseStream.MoveNext(cancellationToken: cancellationToken))
			{
				rezultat.Add(call.ResponseStream.Current);
			}

			Console.WriteLine($"Uspešno pribavljena količine pm2.5 čestica za period izmedju {period.Pocetak} i {period.Kraj}.");
			return Ok(rezultat.Average(p => p.Pm25));
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Greška: {ex.Message}");
			if (ex.Message.Length > 6 && ex.Message.Substring(0, 6) == "Status")
			{
				return ParseStatus(ex.Message);
			}
			return BadRequest(ex.Message);
		}
	}

	[HttpPost("SumPm2_5")]
	public async Task<IActionResult>SumPm2_5([FromBody]Gateway.GrpcClient.VremenskiPeriod period)
	{
		Console.WriteLine($"Pribavljanje sume količine pm2.5 čestica u periodu izmedju {period.Pocetak} i {period.Kraj}");
		try
		{
			List<Gateway.GrpcClient.SenzorPodatak> rezultat = new();
			using var call = client.SviSenzorPodaciPeriod(period);

			while (await call.ResponseStream.MoveNext(cancellationToken: cancellationToken))
			{
				rezultat.Add(call.ResponseStream.Current);
			}

			Console.WriteLine($"Uspešno pribavljena suma količine pm2.5 čestica za period izmedju {period.Pocetak} i {period.Kraj}.");
			return Ok(rezultat.Sum(p => p.Pm25));
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Greška: {ex.Message}");
			if (ex.Message.Length > 6 && ex.Message.Substring(0, 6) == "Status")
			{
				return ParseStatus(ex.Message);
			}
			return BadRequest(ex.Message);
		}
	}

////////////////////////////////////////////

	[HttpPost("MinPm10")]
	public async Task<IActionResult>MinPm10([FromBody]Gateway.GrpcClient.VremenskiPeriod period)
	{
		Console.WriteLine($"Pribavljanje minimalne količine pm10 čestica u periodu izmedju {period.Pocetak} i {period.Kraj}");
		try
		{
			List<Gateway.GrpcClient.SenzorPodatak> rezultat = new();
			using var call = client.SviSenzorPodaciPeriod(period);

			while (await call.ResponseStream.MoveNext(cancellationToken: cancellationToken))
			{
				rezultat.Add(call.ResponseStream.Current);
			}

			Console.WriteLine($"Uspešno pribavljena minimalana količina pm10 čestica za period izmedju {period.Pocetak} i {period.Kraj}.");
			return Ok(rezultat.Min(p => p.Pm10));
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Greška: {ex.Message}");
			if (ex.Message.Length > 6 && ex.Message.Substring(0, 6) == "Status")
			{
				return ParseStatus(ex.Message);
			}
			return BadRequest(ex.Message);
		}
	}

	[HttpPost("MaxPm10")]
	public async Task<IActionResult>MaxPm10([FromBody]Gateway.GrpcClient.VremenskiPeriod period)
	{
		Console.WriteLine($"Pribavljanje maksimalne količine pm10 čestica u periodu izmedju {period.Pocetak} i {period.Kraj}");
		try
		{
			List<Gateway.GrpcClient.SenzorPodatak> rezultat = new();
			using var call = client.SviSenzorPodaciPeriod(period);

			while (await call.ResponseStream.MoveNext(cancellationToken: cancellationToken))
			{
				rezultat.Add(call.ResponseStream.Current);
			}

			Console.WriteLine($"Uspešno pribavljena količine pm10 čestica za period izmedju {period.Pocetak} i {period.Kraj}.");
			return Ok(rezultat.Max(p => p.Pm10));
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Greška: {ex.Message}");
			if (ex.Message.Length > 6 && ex.Message.Substring(0, 6) == "Status")
			{
				return ParseStatus(ex.Message);
			}
			return BadRequest(ex.Message);
		}
	}

	[HttpPost("AvgPm10")]
	public async Task<IActionResult>AvgPm10([FromBody]Gateway.GrpcClient.VremenskiPeriod period)
	{
		Console.WriteLine($"Pribavljanje prosečne količine pm10 čestica u periodu izmedju {period.Pocetak} i {period.Kraj}");
		try
		{
			List<Gateway.GrpcClient.SenzorPodatak> rezultat = new();
			using var call = client.SviSenzorPodaciPeriod(period);

			while (await call.ResponseStream.MoveNext(cancellationToken: cancellationToken))
			{
				rezultat.Add(call.ResponseStream.Current);
			}

			Console.WriteLine($"Uspešno pribavljena količine pm10 čestica za period izmedju {period.Pocetak} i {period.Kraj}.");
			return Ok(rezultat.Average(p => p.Pm10));
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Greška: {ex.Message}");
			if (ex.Message.Length > 6 && ex.Message.Substring(0, 6) == "Status")
			{
				return ParseStatus(ex.Message);
			}
			return BadRequest(ex.Message);
		}
	}

	[HttpPost("SumPm10")]
	public async Task<IActionResult>SumPm10([FromBody]Gateway.GrpcClient.VremenskiPeriod period)
	{
		Console.WriteLine($"Pribavljanje sume količine pm10 čestica u periodu izmedju {period.Pocetak} i {period.Kraj}");
		try
		{
			List<Gateway.GrpcClient.SenzorPodatak> rezultat = new();
			using var call = client.SviSenzorPodaciPeriod(period);

			while (await call.ResponseStream.MoveNext(cancellationToken: cancellationToken))
			{
				rezultat.Add(call.ResponseStream.Current);
			}

			Console.WriteLine($"Uspešno pribavljena suma količine pm10 čestica za period izmedju {period.Pocetak} i {period.Kraj}.");
			return Ok(rezultat.Sum(p => p.Pm10));
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Greška: {ex.Message}");
			if (ex.Message.Length > 6 && ex.Message.Substring(0, 6) == "Status")
			{
				return ParseStatus(ex.Message);
			}
			return BadRequest(ex.Message);
		}
	}

////////////////////////////////////////////////////////////

	private IActionResult ParseStatus(string status)
	{
        string[] tmp = status.Substring(7, status.Length - 8).Split(',');
		if (tmp.Length < 2)
		{
			return BadRequest(status);
		}

        string[] statusCode = tmp[0].Split('=');
        string[] poruka = tmp[1].Split('=');
        if (statusCode[1] == "\"NotFound\"") return NotFound(poruka[1]);
        return BadRequest(status);
	}
}
