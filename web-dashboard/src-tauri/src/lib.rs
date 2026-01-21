#[cfg(desktop)]
mod desktop;

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    #[allow(unused_mut)]
    let mut builder = tauri::Builder::default()
        .setup(|app| {
            if cfg!(debug_assertions) {
                app.handle().plugin(
                    tauri_plugin_log::Builder::default()
                        .level(log::LevelFilter::Info)
                        .build(),
                )?;
            }

            app.handle().plugin(tauri_plugin_notification::init())?;

            #[cfg(desktop)]
            desktop::setup(app)?;

            Ok(())
        });

    #[cfg(desktop)]
    {
        builder = builder.on_window_event(|window, event| {
            desktop::on_window_event(window, event);
        });
    }

    builder
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
