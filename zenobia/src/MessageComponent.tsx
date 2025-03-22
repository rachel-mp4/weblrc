import {message} from "./store"

export default function MessageComponent(message: message) {
    return (<div><b>{message.name}</b> {message.active && "is typing"} { message.text}</div>)
}