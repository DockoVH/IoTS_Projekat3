const tabela = document.getElementById('tabela')

const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
const socket = new WebSocket(`${protocol}//${window.location.host}/ws`);
socket.binaryType = 'arraybuffer';

const temperaturaGranica = 28.0 //60.0
const vlaznostGranica = 60.0 //80.0
const pm2_5Granica = 5.0 //55.0
const pm10Granica = 10.0 //253.0

socket.onopen = (e) => {
    console.log('povezan')
} 

socket.onmessage = (e) => {
    const data = new Uint8Array(e.data)
    let str = ""

    data.forEach(p => {
        str += String.fromCharCode(parseInt(p))
    })

    const obj = JSON.parse(str.slice(1))
    const opts = {
        day: '2-digit',
        year: 'numeric',
        month: '2-digit',
        hour: 'numeric',
        minute: 'numeric',
        second: 'numeric',
        hour12: false
}

    const red = document.createElement('tr')

    const id = document.createElement('td')
    const vreme = document.createElement('td')
    const temp = document.createElement('td')
    const vlaznost = document.createElement('td')
    const pm2_5 = document.createElement('td')
    const pm10 = document.createElement('td')
    
    id.innerText = `${obj.Id}`

    temp.innerText = `${obj.Temperatura.toFixed(2)} °C`
    if (obj.Temperatura > temperaturaGranica) {
        temp.classList.add('prevelika-vrednost')
    }
    vreme.innerText = new Date(obj.Vreme).toLocaleDateString('en-GB', opts)
    vlaznost.innerText = `${obj.Vlaznost.toFixed(2)} %`
    if (obj.Vlaznost > vlaznostGranica) {
        vlaznost.classList.add('prevelika-vrednost')
    }

    pm2_5.innerText = `${obj.Pm2_5.toFixed(2)} µg/m³`
    if (obj.Pm2_5 > pm2_5Granica) {
        pm2_5.classList.add('prevelika-vrednost')
    }

    pm10.innerText = `${obj.Pm10.toFixed(2)} µg/m³`
    if (obj.Pm10 > pm10Granica) {
        pm10.classList.add('prevelika-vrednost')
    }

    red.appendChild(id)
    red.appendChild(vreme)
    red.appendChild(temp)
    red.appendChild(vlaznost)
    red.appendChild(pm2_5)
    red.appendChild(pm10)

    tabela.appendChild(red)
}
