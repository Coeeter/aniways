import type Artplayer from 'artplayer';
import type { components } from '$lib/api/openapi';
import { buttonVariants } from '$lib/components/ui/button';

type StreamInfo = components['schemas']['models.StreamingDataResponse'];

export const thumbnailPlugin = (url: string, getUrl: (path: string) => string) => {
	return (art: Artplayer) => {
		const {
			template: { $progress },
		} = art;

		let timer: NodeJS.Timeout | null = null;
		const abortController = new AbortController();

		if (!url) return { name: 'thumbnailPlugin' };

		art.on('destroy', () => {
			abortController.abort();
			if (timer) clearTimeout(timer);
		});

		fetch(url, { signal: abortController.signal })
			.then((res) => res.text())
			.then((res) => {
				const tns = res
					.split('\n')
					.filter((line) => line.trim())
					.slice(1);

				const data: {
					start: number;
					end: number;
					url: string;
					x: number;
					y: number;
					w: number;
					h: number;
				}[] = [];

				tns.forEach((_, index) => {
					if (index % 3 !== 0) return;
					const time = tns[index + 1];
					const spriteUrl = tns[index + 2];
					if (!time || !spriteUrl) return;
					const start = time.split(' --> ')[0]!;
					const end = time.split(' --> ')[1]!;

					const startSeconds = start.split(':').reduce((acc, time, i) => {
						return acc + Number(time) * Math.pow(60, 2 - i);
					}, 0);

					const endSeconds = end.split(':').reduce((acc, time, i) => {
						return acc + Number(time) * Math.pow(60, 2 - i);
					}, 0);

					const [x, y, w, h] = spriteUrl.split('#xywh=')[1]!.split(',').map(Number);
					const spritePath = spriteUrl.split('#xywh=')[0];

					data.push({
						start: startSeconds,
						end: endSeconds,
						url: spritePath ? getUrl(spritePath) : '',
						x: x!,
						y: y!,
						w: w!,
						h: h!,
					});
				});

				// Preload images for smoother experience
				data
					.map((item) => item.url)
					.filter((src) => src)
					.forEach((src) => {
						const img = new Image();
						img.src = src;
					});

				art.controls.add({
					name: 'vtt-thumbnail',
					position: 'top',
					mounted($control) {
						$control.classList.add('art-control-thumbnails');
						art.on('setBar', async (type, percentage, event) => {
							const isMobile =
								/Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(
									navigator.userAgent,
								);

							const isMobileDragging = type === 'played' && event && isMobile;

							if (type === 'hover' || isMobileDragging) {
								const width = $progress.clientWidth * percentage;
								const second = percentage * art.duration;
								$control.style.display = 'flex';

								const thumbnail = data.find((item) => item.start <= second && item.end >= second);

								if (!thumbnail) {
									$control.style.display = 'none';
									return;
								}

								if (width > 0 && width < $progress.clientWidth) {
									$control.style.backgroundImage = `url(${thumbnail.url})`;
									$control.style.height = `${thumbnail.h}px`;
									$control.style.width = `${thumbnail.w}px`;
									$control.style.backgroundPosition = `-${thumbnail.x}px -${thumbnail.y}px`;
									if (width <= thumbnail.w / 2) {
										$control.style.left = '0px';
									} else if (width > $progress.clientWidth - thumbnail.w / 2) {
										$control.style.left = `${$progress.clientWidth - thumbnail.w}px`;
									} else {
										$control.style.left = `${width - thumbnail.w / 2}px`;
									}
								} else {
									if (!isMobile) {
										$control.style.display = 'none';
									}
								}

								if (isMobileDragging) {
									if (timer) clearTimeout(timer);
									timer = setTimeout(() => {
										$control.style.display = 'none';
									}, 1000);
								}
							}
						});
					},
				});
			})
			.catch((error) => {
				if (error instanceof Error && error.name === 'AbortError') {
					return;
				}
				console.error('Thumbnail plugin error:', error);
			});

		return { name: 'thumbnailPlugin' };
	};
};

export const skipPlugin = (source: StreamInfo) => {
	return (art: Artplayer) => {
		art.on('ready', () => {
			function addElement(title: string, start: number, end: number) {
				const startPercentage = (start / art.duration) * 100;
				const endPercentage = ((end - start) / art.duration) * 100;
				const highlightElement = art.template.$progress.querySelector('.art-progress-highlight');

				highlightElement?.insertAdjacentHTML(
					'beforeend',
					`<span data-time="${start}" data-text="${title}" style="left: ${startPercentage}%; width: ${endPercentage}% !important"></span>`,
				);
			}

			if (source.intro) {
				addElement('Opening', source.intro.start, source.intro.end);
			}

			if (source.outro) {
				addElement('Ending', source.outro.start, source.outro.end);
			}
		});

		art.on('video:timeupdate', () => {
			if (
				source.intro &&
				art.currentTime >= source.intro.start &&
				art.currentTime <= source.intro.end &&
				!art.layers['opening']
			) {
				const height = art.template.$controls.getBoundingClientRect().height;
				art.layers.add({
					name: 'opening',
					position: 'top',
					html: `<button class="${buttonVariants({ class: 'absolute right-[10px] pointer-events-auto touch-manipulation' })}" style="bottom: ${height + 20}px;">Skip Opening</button>`,
					click: (_, e) => {
						e.preventDefault();
						e.stopPropagation();

						if (e instanceof TouchEvent) {
							e.stopImmediatePropagation();
						}

						art.seek = source.intro!.end;
						art.notice.show = 'Skipped Opening';
					},
				});
			}

			if (
				source.outro &&
				art.currentTime >= source.outro.start &&
				art.currentTime <= source.outro.end &&
				!art.layers['ending']
			) {
				const height = art.template.$controls.getBoundingClientRect().height;
				art.layers.add({
					name: 'ending',
					position: 'top',
					html: `<button class="${buttonVariants({ class: 'absolute right-[10px] pointer-events-auto touch-manipulation' })}" style="bottom: ${height + 20}px;">Skip Ending</button>`,
					click: (_, e) => {
						e.preventDefault();
						e.stopPropagation();

						if (e instanceof TouchEvent) {
							e.stopImmediatePropagation();
						}

						art.seek = source.outro!.end;
						art.notice.show = 'Skipped Ending';
					},
				});
			}

			if (
				source.intro &&
				(art.currentTime < source.intro.start || art.currentTime > source.intro.end) &&
				art.layers['opening']
			) {
				art.layers.remove('opening');
			}

			if (
				source.outro &&
				(art.currentTime < source.outro.start || art.currentTime > source.outro.end) &&
				art.layers['ending']
			) {
				art.layers.remove('ending');
			}
		});

		return { name: 'skipPlugin' };
	};
};

export const windowKeyBindPlugin = () => {
	return (art: Artplayer) => {
		const keydownHandler = (e: Event) => {
			if (e instanceof KeyboardEvent === false) return;
			if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return;

			if (Object.keys(art.hotkey.keys).includes(e.code)) {
				e.preventDefault();
				e.stopPropagation();
				art.hotkey.keys[e.code]?.forEach((fn) => fn?.(e));
			}
		};

		art.events.proxy(window, 'keydown', keydownHandler);

		art.on('destroy', () => {
			window.removeEventListener('keydown', keydownHandler);
		});

		art.on('ready', () => {
			art.hotkey.add('KeyF', () => {
				art.fullscreen = !art.fullscreen;
			});

			art.hotkey.add('KeyM', () => {
				art.muted = !art.muted;
			});

			art.hotkey.add('Space', () => {
				art.toggle();
			});

			art.hotkey.add('ArrowLeft', () => {
				art.backward = 10;
			});

			art.hotkey.add('ArrowRight', () => {
				art.forward = 10;
			});

			art.hotkey.add('ArrowUp', () => {
				art.volume += 0.1;
			});

			art.hotkey.add('ArrowDown', () => {
				art.volume -= 0.1;
			});
		});

		return { name: 'windowKeyBindPlugin' };
	};
};

export const amplifyVolumePlugin = () => {
	return (art: Artplayer) => {
		let context: AudioContext | null = null;
		let source: MediaElementAudioSourceNode | null = null;
		let gainNode: GainNode | null = null;

		art.on('ready', () => {
			art.volume = 100;

			context = new AudioContext();
			source = context.createMediaElementSource(art.video);
			gainNode = context.createGain();
			source.connect(gainNode);
			gainNode.connect(context.destination);
			gainNode.gain.value = 2;

			art.on('video:play', () => {
				context?.resume();
			});

			art.on('video:pause', () => {
				context?.suspend();
			});
		});

		art.on('destroy', () => {
			if (source) {
				source.disconnect();
				source = null;
			}
			if (gainNode) {
				gainNode.disconnect();
				gainNode = null;
			}
			if (context && context.state !== 'closed') {
				context.close();
				context = null;
			}
		});

		return { name: 'amplifyVolumePlugin' };
	};
};
