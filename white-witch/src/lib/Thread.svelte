<script lang="ts">
    import MessageComponent from "./MessageComponent.svelte";
    import { messages } from "../store"
</script>

<div>
{#each $messages as message (message.id)}
    <MessageComponent {message} />
{/each}
</div>
