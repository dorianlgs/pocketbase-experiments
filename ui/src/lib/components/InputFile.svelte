<script>
	let { errors, name, placeholder } = $props();
</script>

<input
	id={name}
	{name}
	type={'file'}
	{placeholder}
	class="{errors[name] ? 'input-error' : ''} input-md"
/>

<style>
	/* file upload button */
	input[type='file']::file-selector-button {
		border-radius: 4px;
		padding: 0 16px;
		height: 40px;
		cursor: pointer;
		background-color: white;
		border: 1px solid rgba(0, 0, 0, 0.16);
		box-shadow: 0px 1px 0px rgba(0, 0, 0, 0.05);
		margin-right: 16px;
		transition: background-color 200ms;
	}

	/* file upload button hover state */
	input[type='file']::file-selector-button:hover {
		background-color: #f3f4f6;
	}

	/* file upload button active state */
	input[type='file']::file-selector-button:active {
		background-color: #e5e7eb;
	}
</style>
