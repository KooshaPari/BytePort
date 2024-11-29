<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Button from '$lib/components/ui/button';
	import Icon from '@iconify/svelte';
	import type { Project, Instance } from '$lib/utils';
	import GitSearch from './gitSearch.svelte';
	import { onMount, onDestroy } from 'svelte';
	import type { User } from '../stores/user';
	import type { Repository } from '$lib/git';
	let client: User | null = null;
	let clientRepo: Repository | null = null;
	let projheadTxt: string = 'Welcome.';
	let stage = 1;
	let projHeadDescr: string = "Let's Begin First Time Setup";
	let newProject: Project = {
		UUID: '',
		Owner: '',
		User: client,
		Name: '',
		Repository: clientRepo,
		Description: '',
		LastUpdated: '',
		Status: '',
		Type: '',
		Instances: []
	};

	function addNewProject() {
		// open dialog
		stage = 1;
		setStage();
		return;
	}
	async function setStage() {
		if (stage === 1) {
			stage++;
		}
		const projForm = document.querySelector(`#form-stage-${stage}`) as HTMLFormElement;
		stage++;
		switch (stage) {
			case 1:
				projheadTxt = "Let's Start With Some Basic Information...";
				projHeadDescr = 'Enter Your AWS Credentials Below';
				break;
			case 2:
				projheadTxt = "Let's Continue With OpenAI Credentials...";
				projHeadDescr = 'Enter Your OpenAI Credentials Below';
				break;
			case 3:
				projheadTxt = "Let's Connect Your Portfolio...";
				projHeadDescr = 'Provide Your Portfolios';
				break;
			case 4:
				projheadTxt = "Let's Connect Your Git Provider";
				projHeadDescr = 'Please Continue On GitHub';

				break;
			case 5:
				projheadTxt = 'Setup Complete!';
				projHeadDescr = 'You have completed the first-time setup.';
				break;
			default:
				projheadTxt = 'Setup ERR!';
				break;
		}
	}
</script>

<GitSearch />
