use enigo::Keyboard;
use std::{collections::VecDeque, sync::Arc};
use tokio::sync::Mutex;
use tokio::time::{sleep, Duration};
use tokio_util::sync::CancellationToken;

use tokio::task;

#[derive(Debug, Clone)]
pub struct ActionSequencerAction {
    pub keys: String,
    pub press_time: Option<f32>, /* if none specified - key will be held */
    pub wait_time: Option<f32>,
    pub release: Option<bool>, /* if specified and set to true key will be released */
}

pub struct ActionSequencer {
    pub actions: Arc<Mutex<VecDeque<ActionSequencerAction>>>,
    enigo: Arc<Mutex<enigo::Enigo>>,
}

impl ActionSequencer {
    pub fn new() -> ActionSequencer {
        ActionSequencer {
            actions: Arc::new(Mutex::new(VecDeque::new())),
            enigo: Arc::new(Mutex::new(
                enigo::Enigo::new(&enigo::Settings::default()).unwrap(),
            )),
        }
    }

    pub async fn add_action(&self, action: ActionSequencerAction) {
        self.actions.lock().await.push_back(action);
    }

    pub fn parse_keys<T: AsRef<str>>(input: T) -> (Vec<enigo::Key>, Vec<enigo::Key>) {
        use enigo::Key;
        let split = input.as_ref().split('+');
        let modifier_keys = split
            .clone()
            .filter_map(|key| match key.to_lowercase().as_str() {
                "ctrl" | "control" => Some(Key::Control),
                "alt" => Some(Key::Alt),
                "meta" | "cmd" | "command" => Some(Key::Meta),
                _ => None,
            })
            .collect();

        let action_keys: Vec<enigo::Key> = split
            .clone()
            .filter_map(|key| match key.to_lowercase().as_str() {
                "shift" => Some(Key::Shift),
                "backspace" => Some(Key::Backspace),
                "delete" => Some(Key::Delete),
                "arrowdown" | "down" => Some(Key::DownArrow),
                "arrowup" | "up" => Some(Key::UpArrow),
                "arrowleft" | "left" => Some(Key::LeftArrow),
                "arrowright" | "right" => Some(Key::RightArrow),
                "return" | "enter" => Some(Key::Return),
                "space" | "spacebar" => Some(Key::Space),
                "tab" => Some(Key::Tab),
                "escape" | "esc" => Some(Key::Escape),
                "capslock" => Some(Key::CapsLock),
                "f1" => Some(Key::F1),
                "f2" => Some(Key::F2),
                "f3" => Some(Key::F3),
                "f4" => Some(Key::F4),
                "f5" => Some(Key::F5),
                "f6" => Some(Key::F6),
                "f7" => Some(Key::F7),
                "f8" => Some(Key::F8),
                "f9" => Some(Key::F9),
                "f10" => Some(Key::F10),
                "f11" => Some(Key::F11),
                "f12" => Some(Key::F12),
                "pageup" => Some(Key::PageUp),
                "pagedown" => Some(Key::PageDown),
                "home" => Some(Key::Home),
                "end" => Some(Key::End),
                "insert" => Some(Key::Insert),
                key if key.len() == 1 => {
                    let char = key.chars().next().unwrap();
                    Some(Key::Unicode(char))
                }
                _ => None,
            })
            .collect();

        (modifier_keys, action_keys)
    }

    pub async fn press_or_release_keys<T: AsRef<str>>(
        enigo: Arc<Mutex<enigo::Enigo>>,
        keys: T,
        direction: enigo::Direction,
    ) {
        let mut enigo_lock = enigo.lock().await;
        let (modifier_keys, action_keys) = ActionSequencer::parse_keys(keys.as_ref());

        if direction == enigo::Direction::Press {
            for key in modifier_keys.iter() {
                enigo_lock.key(*key, enigo::Direction::Press).unwrap();
            }
            if modifier_keys.is_empty() != true {
                sleep(Duration::from_millis(30)).await;
            }
            for key in action_keys.iter() {
                enigo_lock.key(*key, enigo::Direction::Press).unwrap();
            }
        } else {
            for key in action_keys.iter() {
                enigo_lock.key(*key, enigo::Direction::Release).unwrap();
            }
            if modifier_keys.is_empty() != true {
                sleep(Duration::from_millis(30)).await;
            }
            for key in modifier_keys.iter() {
                enigo_lock.key(*key, enigo::Direction::Release).unwrap();
            }
        }
    }

    pub fn run(&self, cancel_token: CancellationToken) -> task::JoinHandle<()> {
        let enigo_arc: Arc<Mutex<enigo::Enigo>> = Arc::clone(&self.enigo);
        let actions_queue = Arc::clone(&self.actions);
        let thread = task::spawn(async move {
            loop {
                tokio::select! {
                  _ = cancel_token.cancelled() => {
                      break;
                  }
                  _ = async {
                    while let Some(action) = actions_queue.lock().await.pop_front() {
                      match action.release {
                          Some(true) => {
                            ActionSequencer::press_or_release_keys(Arc::clone(&enigo_arc), &action.keys, enigo::Direction::Release).await;
                          },
                          Some(false) | None => match action.press_time {
                              Some(press_time) => {
                                ActionSequencer::press_or_release_keys(Arc::clone(&enigo_arc), &action.keys, enigo::Direction::Press).await;
                                sleep(Duration::from_millis((press_time * 1000.0).abs() as u64)).await;
                                ActionSequencer::press_or_release_keys(Arc::clone(&enigo_arc), &action.keys, enigo::Direction::Release).await;
                                sleep(Duration::from_millis((action.wait_time.unwrap_or(0.1) * 1000.0).abs() as u64)).await;
                              },
                              None =>{
                                ActionSequencer::press_or_release_keys(Arc::clone(&enigo_arc), &action.keys, enigo::Direction::Press).await;
                              },
                          },
                      }
                  }
                  } => {}
                }
            }
        });
        thread
    }
}
