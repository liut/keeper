//
// Package reaper used for loop task 用定时循环任务
//
package reaper

//
// Example: default interval is 5 minutes
// defer reaper.Quit(reaper.Run(0, func() error {
// 		// Loop process
// }))
//

// // for start
// quit, done := reaper.Run(0, func() error {
// 	// Loop process
// })
//
// // for stop
// reaper.Quit(quit, done)
