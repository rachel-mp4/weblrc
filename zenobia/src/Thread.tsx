import { messages } from "./store"
import MessageComponent from "./MessageComponent"

export default function Thread() {
    return (
        <div>{messages.value.map(msg => (
            MessageComponent(msg)
        ))}</div>
    )
}