<script lang="ts">
    import { checkHealth } from "./api/health"

    type Status = "idle" | "checking" | "success" | "failed"

    let status = $state<Status>("idle")
    let detail = $state("")

    const handleHealth = async () => {
        // Ignore clicks while a check is already in flight.
        if (status === "checking") return

        status = "checking"
        detail = ""

        const result = await checkHealth()

        if (result.ok) {
            status = "success"
            detail = result.status
        }
        else {
            status = "failed"
            detail = result.error
        }
    }
</script>

<div>helath route</div>
<button onclick={handleHealth} disabled={status === "checking"}>{status}</button>
{#if detail}
    <p>{detail}</p>
{/if}
