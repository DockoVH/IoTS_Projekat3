using System;

namespace Gateway.Models;

public class SenzorPodatak
{
	public int ID { get; set; }
	public DateTime Vreme { get; set; }
	public float Temperatura { get; set; }
	public float VlaznostVazduha { get; set; }
	public float Pm2_5 { get; set; }
	public float Pm10 { get; set; }
}
