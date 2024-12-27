<script lang="ts">
  import { superForm } from 'sveltekit-superforms/client';
  import { zodClient } from 'sveltekit-superforms/adapters';
  import type { SuperValidated } from 'sveltekit-superforms';
  import { formSchema } from './schema';

  // Define default form values
  const defaultFormValues = {
    github: '',
    aws: { 
      accessKey: '', 
      secretKey: '' 
    },
    openai: '',
    demo: { 
      endpoint: '', 
      apiKey: '' 
    },
    modal: 'local' as const
  };

  export let data: { form?: SuperValidated<typeof formSchema> };
  export let item: string;

  // Initialize form with default values if data.form is undefined
  const { form, validate, enhance } = superForm(
    data?.form ?? defaultFormValues,
    {
      validators: zodClient(formSchema),
      dataType: 'json',
      // Ensure form is valid before submitting
      onSubmit: ({ form, data }) => {
        // Optional: Add any pre-submission validation here
        return { form, data };
      },
      // Handle validation errors
      onError: ({ result }) => {
        console.error('Form validation failed:', result);
      }
    }
  );
</script>

<form method="POST" class="space-y-8" use:enhance>
  {#if item === 'integrations'}
    <div class="openAICard align-center flex flex-row gap-3">
      <!-- Wrap form fields in #if $form -->
      {#if $form}
        <Form.Field class=" " name="github" form={$form}>
          <Form.Control let:attrs>
            <Form.Label>Github Integration</Form.Label>
            <Input {...attrs} bind:value={$form.github} />
          </Form.Control>
          <Form.FieldErrors />
        </Form.Field>
        
        <!-- Rest of your form fields here -->
      {/if}
    </div>
  {/if}

  <Form.Button>Save {item}</Form.Button>
</form>

{#if browser}
  <SuperDebug data={$form} />
{/if}