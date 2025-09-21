const tempPolje = document.getElementById('temperatura')
const vlaznostPolje = document.getElementById('vlaznost-vazduha')
const pm2_5Polje = document.getElementById('pm2_5')
const pm10olje = document.getElementById('pm10')

const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
const socket = new WebSocket(`${protocol}//${window.location.host}/ws`);
socket.binaryType = 'arraybuffer';

document.onbeforeunload = (e) => {
	socket.send("GOTOVO")
}

socket.onopen = (e) => {
    console.log('povezan')
} 

socket.onmessage = (e) => {
    const data = new Uint8Array(e.data)
    let str = ""

    data.forEach(p => {
        str += String.fromCharCode(parseInt(p))
    })

    console.log(`primljena poruka ${str}`)

    const obj = JSON.parse(str.slice(1))
    
    const novoPolje = document.createElement('div')
    novoPolje.classList.add('polje')

    switch (str[0])
    {
        case '0':
            novoPolje.innerText = `ID: ${obj.Id}, temperatura: ${obj.Temperatura}`
            tempPolje.appendChild(novoPolje)
            break
        case '1':
            novoPolje.innerText = `ID: ${obj.Id}, Vlažnost vazduha: ${obj.Vlaznost}`
            vlaznostPolje.appendChild(novoPolje)
            break;
        case '2':
            novoPolje.innerText = `ID: ${obj.Id}, Pm2.5: ${obj.Pm2_5}`
            pm2_5Polje.appendChild(novoPolje)
            break
        case '3':
            novoPolje.innerText = `ID: ${obj.Id}, Pm10: ${obj.Pm10}`
            pm10olje.appendChild(novoPolje)
            break;
        default:
            break
    }
}
