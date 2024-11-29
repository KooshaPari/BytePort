<Popover.Content class="w-[200px] p-0">
			<Command.Root>
				<Command.Input placeholder="Search framework..." />
				<Command.Empty>No framework found.</Command.Empty>
				<Command.Group>
					{#each frameworks as framework}
						<Command.Item
							value={framework.value}
							onSelect={(currentValue) => {
								value = currentValue;
								closeAndFocusTrigger(ids.trigger);
							}}
						>
							<Check class={cn('mr-2 h-4 w-4', value !== framework.value && 'text-transparent')} />
							{framework.label}
						</Command.Item>
					{/each}
				</Command.Group>
			</Command.Root>
		</Popover.Content>
	</Popover.Root>
	<Input name="nvms" />