//
// Package reaper 定时任务，用于资源清理
//
package reaper

//
// Example: default interval is 5 minutes
// defer reaper.Quit(reaper.Run(0, func() error {
// 		// Cleanup
// }))
//
