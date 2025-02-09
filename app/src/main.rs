#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")] // hide console window on Windows in release
#![allow(rustdoc::missing_crate_level_docs)] // it's an example

use clap::{Parser, Subcommand};
use direct_controller::DirectControlCommand;
use std::sync::Arc;

use controller_manager::ControllerManagerChangeEvent;
use eframe::egui;
use tokio::sync::Mutex;
use tokio_util::sync::CancellationToken;

mod action_sequencer;
mod commands;
mod config_defs;
mod config_loader;
mod controller_manager;
mod direct_controller;
mod profile_runner;

#[derive(Subcommand, Debug, Clone)]
enum Commands {
    Calibrate {
        #[arg(short, long, default_value = ".config")]
        config_dir: String,
    },
}

#[derive(Parser, Debug)]
#[command(version, about, long_about = None)]
struct Args {
    #[command(subcommand)]
    cmd: Option<Commands>,
}

#[tokio::main]
async fn main() -> eframe::Result {
    env_logger::init(); // Log to stderr (if you run with `RUST_LOG=debug`).

    let args = Args::parse();
    match args.cmd {
        Some(Commands::Calibrate { config_dir }) => {
            commands::run_calibration_mode::run_calibration_mode(config_dir).await;
            return Ok(());
        }
        None => {
            println!("No command provided - running UI");
        }
    }

    let cancel_token = CancellationToken::new();

    let (on_selected_profile_change_sender, mut on_selected_profile_change_receiver) =
        tokio::sync::watch::channel::<Option<String>>(None);

    let mut config = config_loader::ConfigLoader::new();
    config.load_from_dir(Some(".config"));
    config.load_from_dir(Some("config"));
    let shared_config = Arc::new(config);

    let sequencer = Arc::new(action_sequencer::ActionSequencer::new());

    let (direct_controller_sender, direct_controller_receiver) =
        tokio::sync::broadcast::channel::<DirectControlCommand>(10000);
    let direct_controller = direct_controller::DirectController::new().await;

    let profile_runner = Arc::new(Mutex::new(profile_runner::ProfileRunner::new(
        Arc::clone(&shared_config),
        Arc::clone(&sequencer),
        Arc::new(Mutex::new(direct_controller_sender)),
    )));

    let (controller_manager_event_channel_sender, mut controller_manager_event_channel_receiver) =
        tokio::sync::broadcast::channel::<ControllerManagerChangeEvent>(10000);

    let controller_manager_config: Arc<config_loader::ConfigLoader> = Arc::clone(&shared_config);
    let controller_manager_cancel_token = cancel_token.clone();
    tokio::task::spawn_blocking(move || {
        let mut controller_manager =
            controller_manager::ControllerManager::new(controller_manager_config);
        controller_manager.subscribe(
            controller_manager_event_channel_sender,
            controller_manager_cancel_token.clone(),
        );
        controller_manager.attach(controller_manager_cancel_token.clone());
    });

    let profile_listener_cancel_token = cancel_token.clone();
    let profile_listener_profile_runner_clone = Arc::clone(&profile_runner);
    tokio::task::spawn(async move {
        loop {
            tokio::select! {
                _ = profile_listener_cancel_token.cancelled() => {
                    break;
                },
                _ = on_selected_profile_change_receiver.changed() => {
                    let profile = on_selected_profile_change_receiver.borrow().clone();
                    match profile {
                        Some(profile) => {
                            println!("Selected profile: {}", profile.clone());
                            profile_listener_profile_runner_clone.lock().await.set_profile(profile).unwrap();
                        },
                        None => {
                            println!("Cleared Profile");
                            profile_listener_profile_runner_clone.lock().await.reset_profile().unwrap();
                        }
                    }
                }
            }
        }
    });

    let event_listener_cancel_token = cancel_token.clone();
    tokio::task::spawn(async move {
        loop {
            tokio::select! {
                _ = event_listener_cancel_token.cancelled() => {
                    break;
                }
                _ = async {
                    let event = controller_manager_event_channel_receiver.recv().await.unwrap();
                    profile_runner.lock().await.run(event).await;
                } => {}
            }
        }
    });

    sequencer.run(cancel_token.clone());
    direct_controller.start(
        cancel_token.clone(),
        Arc::new(Mutex::new(direct_controller_receiver)),
    );

    let options = eframe::NativeOptions {
        viewport: egui::ViewportBuilder::default().with_inner_size([400.0, 300.0]),
        ..Default::default()
    };
    eframe::run_native(
        "TSW5 Throttle Mapper",
        options,
        Box::new(|_| {
            Ok(Box::new(MainApp {
                config: shared_config,
                ui_close_token: cancel_token,
                selected_profile: None,
                on_selected_profile_change_sender,
            }))
        }),
    )
}

struct MainApp {
    config: Arc<config_loader::ConfigLoader>,
    ui_close_token: CancellationToken,

    /* local state */
    selected_profile: Option<String>,

    /* channels */
    on_selected_profile_change_sender: tokio::sync::watch::Sender<Option<String>>,
}

impl eframe::App for MainApp {
    fn update(&mut self, ctx: &egui::Context, _frame: &mut eframe::Frame) {
        let mut selected_profile = self.selected_profile.clone();

        egui::CentralPanel::default().show(ctx, |ui| {
            ui.vertical(|ui| {
                egui::ComboBox::from_label("Select profile")
                    .selected_text(format!(
                        "{}",
                        match &selected_profile {
                            Some(profile) => profile.clone(),
                            None => String::from(""),
                        }
                    ))
                    .show_ui(ui, |ui| {
                        ui.selectable_value(&mut selected_profile, None, "None");
                        for profile in self.config.controller_profiles.iter() {
                            ui.selectable_value(
                                &mut selected_profile,
                                Some(profile.name.clone()),
                                profile.name.clone(),
                            );
                        }
                    });
            });
        });

        if selected_profile != self.selected_profile {
            self.selected_profile = selected_profile.clone();
            self.on_selected_profile_change_sender
                .send(selected_profile.clone())
                .unwrap();
        }
    }

    fn on_exit(&mut self, _gl: Option<&eframe::glow::Context>) {
        self.ui_close_token.cancel();
    }
}
