<script lang="ts">
    import type { Message } from "../store.ts";
    export let message: Message;
</script>

<div>
    <b>{message.name}</b>
    {#if message.active}
        is typing
    {/if}
    {message.text}
</div>
