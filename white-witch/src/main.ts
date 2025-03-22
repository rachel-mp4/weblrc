import { mount } from 'svelte'
import App from './App.svelte'
import { messages, topic } from "./store";

// const ws = new WebSocket("ws://localhost:927/ws");
// ws.binaryType = "arraybuffer";
// ws.onopen = () => {
//   console.log("connected");
// };
// ws.onmessage = (event) => {

//   parseEvent(event);
// };
// ws.onclose = (event) => {
//   console.log("disconnected");
// };

async function fuzz() {
  const start = performance.now()
  let max = 0
  let actives:number[] = []
  for (let i = 0; i < 8192; i++) {
    const r = Math.random()
    if (r < .006) {
      await new Promise(r => setTimeout(r, 20));
    } else if (r < .5) {
      if (actives.length !== 0) {
        const idx = Math.floor(actives.length * Math.random())
        const id = actives[idx]
        fuzzAppend(id)
      }
    } else if (r < .7) {
      fuzzInit(max)
      actives.push(max)
      max += 1
    } else if (r < .9) {
      if (actives.length !== 0) {
        const idx = Math.floor(actives.length * Math.random())
        const id = actives[idx]
        actives = actives.filter((num) => num !== id)
        fuzzPub(id)
      }
    } else {
      fuzzPing()
    }
  }
  requestIdleCallback(() => {
    const end = performance.now()
    console.log(`UI update took ${end - start}ms and we did ${max} messages`);
  })
}

function fuzzPing() {
  const newPing = "" + Math.random()
  const textArray = new TextEncoder().encode(newPing)
  parseEventArray(new Uint8Array([0, 0, 0, 0, 0, 0, ...textArray]))
}

function fuzzInit(id: number) {
  const name = "" + Math.random()
  const textArray = new TextEncoder().encode(name)
  if (id <= 255) {
    parseEventArray(new Uint8Array([0, 0, 0, 0, id, 2, 0, 0, ...textArray]))
  } else {
    const idArray = new Uint8Array(2)
    const view = new DataView(idArray.buffer)
    view.setUint16(0, id, false)
    parseEventArray(new Uint8Array([0, 0, 0, ...idArray, 2, 0, 0, ...textArray]))
  }
}

function fuzzPub(id: number) {
  if (id <= 255) {
    parseEventArray(new Uint8Array([0, 0, 0, 0, id, 3]))
  } else {
    const idArray = new Uint8Array(2)
    const view = new DataView(idArray.buffer)
    view.setUint16(0, id, false)
    parseEventArray(new Uint8Array([0, 0, 0, ...idArray, 3]))
  }
}

function fuzzAppend(id: number) {
  const num = Math.floor(Math.random() * (126 - 32 + 1)) + 32
  if (id <= 255) {
    parseEventArray(new Uint8Array([0, 0, 0, 0, id, 4, 0, 0, num]))
  } else {
    const idArray = new Uint8Array(2)
    const view = new DataView(idArray.buffer)
    view.setUint16(0, id, false)
    parseEventArray(new Uint8Array([0, 0, 0, ...idArray, 4, 0, 0, num]))
  }
}

function parseEventArray(byteArray: Uint8Array) {
  switch (byteArray[5]) {
    case 0: {
      const text = new TextDecoder("ascii").decode(byteArray.slice(6));
      topic.update(() => {
        return text;
      })
      return;
    }

    case 2: {
      const id = readId(byteArray.slice(1, 5));
      const color = byteArray[7];
      const name = new TextDecoder("ascii").decode(byteArray.slice(8));
      const text = "";
      const active = true;
      messages.update((msgs) => {
        return [...msgs, { id, color, name, text, active }];
      });
      return;
    }

    case 3: {
      const id = readId(byteArray.slice(1, 5));
      messages.update((msgs) =>
        msgs.map((msg) =>
          msg.id === id ? { ...msg, active: false } : msg
        )
      );
      return;
    }

    case 4: {
      const id = readId(byteArray.slice(1, 5));
      const idx = readIdx(byteArray.slice(6, 8));
      const s = new TextDecoder("ascii").decode(byteArray.slice(8));
      messages.update((msgs) =>
        msgs.map((msg) =>
          msg.id === id ? { ...msg, text: msg.text.slice(0, idx) + s + msg.text.slice(idx) } : msg
        )
      );
      return;
    }

    case 5: {
      const id = readId(byteArray.slice(1, 5));
      const idx = readIdx(byteArray.slice(6, 8));
      messages.update((msgs) =>
        msgs.map((msg) =>
          msg.id === id ? { ...msg, text: msg.text.slice(0, idx - 1) + msg.text.slice(idx) } : msg
        )
      );
      return;
    }
  }
}

function parseEvent(event: MessageEvent<any>): void {
  const byteArray = new Uint8Array(event.data);
  parseEventArray(byteArray)
}

function readId(bytes: Uint8Array): number {
  return new DataView(
    bytes.buffer,
    bytes.byteOffset,
    bytes.byteLength,
  ).getUint32(0, false);
}

function readIdx(bytes: Uint8Array): number {
  return new DataView(
    bytes.buffer,
    bytes.byteOffset,
    bytes.byteLength,
  ).getUint16(0, false);
}



const app = mount(App, {
  target: document.getElementById('app')!,
})

export default app

await fuzz()