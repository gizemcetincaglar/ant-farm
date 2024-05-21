package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// verileri tutmak için struct oluşturduk
type Room struct {
	Name string
	X    int
	Y    int
}

type Tunnel struct {
	Room1 string
	Room2 string
}

type AntFarm struct {
	AntCount  int
	StartRoom Room
	EndRoom   Room
	Rooms     map[string]Room
	Tunnels   []Tunnel
}

func main() {
	// program ne kadar sürede çlışıyor diye kontrol ediyorum
	start := time.Now()
	if len(os.Args) < 2 {
		fmt.Println("Please specify a file name: go run main.go <filename>")
		return
	}

	fileName := os.Args[1]
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Failed to read file:", err)
		return
	}
	defer file.Close()
	// AntFarm struct'undan bir değişken oluşturdum
	var farm AntFarm
	// bu structun içerisindeki bir değiken ile bir slice oluşturdum
	farm.Rooms = make(map[string]Room)
	var startRoomSet, endRoomSet bool
	// dosya okuma
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if farm.AntCount == 0 {
			antCount, err := strconv.Atoi(line)
			if err != nil || antCount <= 0 {
				fmt.Println("ERROR: invalid data format")
				return
			}
			farm.AntCount = antCount
			continue
		}

		if strings.HasPrefix(line, "##start") {
			if scanner.Scan() {
				startRoomInfo := strings.Fields(scanner.Text())
				if len(startRoomInfo) != 3 {
					fmt.Println("ERROR: invalid data format")
					return
				}
				farm.StartRoom = parseRoom(startRoomInfo)
				if err := validateRoom(farm.StartRoom); err != nil {
					fmt.Println("ERROR:  invalid data format", err)
					return
				}
				farm.Rooms[farm.StartRoom.Name] = farm.StartRoom
				startRoomSet = true
			}
			continue
		}

		if strings.HasPrefix(line, "##end") {
			if scanner.Scan() {
				endRoomInfo := strings.Fields(scanner.Text())
				if len(endRoomInfo) != 3 {
					fmt.Println("ERROR: invalid data format")
					return
				}
				farm.EndRoom = parseRoom(endRoomInfo)
				if err := validateRoom(farm.EndRoom); err != nil {
					fmt.Println("ERROR:", err)
					return
				}
				farm.Rooms[farm.EndRoom.Name] = farm.EndRoom
				endRoomSet = true
			}
			continue
		}
		// yuva bilgileri
		if strings.Contains(line, " ") && !strings.HasPrefix(line, "##") {
			roomInfo := strings.Fields(line)
			if len(roomInfo) != 3 {
				fmt.Println("ERROR: invalid room name format")
				return
			}
			room := parseRoom(roomInfo)
			if err := validateRoom(room); err != nil {
				fmt.Printf("ERROR: Invalid room - %s\n", err)
				return
			}

			if _, exists := farm.Rooms[room.Name]; exists {
				fmt.Println("ERROR: duplicate room names")
				return
			}
			farm.Rooms[room.Name] = room
			continue
		}

		if strings.Contains(line, "-") {
			tunnelInfo := strings.Split(line, "-")
			if len(tunnelInfo) != 2 {
				fmt.Println("ERROR: invalid data format")
				return
			}
			if _, exists := farm.Rooms[tunnelInfo[0]]; !exists {
				fmt.Println("ERROR: unknown room in connection")
				return
			}
			if _, exists := farm.Rooms[tunnelInfo[1]]; !exists {
				fmt.Println("ERROR: unknown room in connection")
				return
			}
			if tunnelInfo[0] == tunnelInfo[1] {
				fmt.Println("ERROR: invalid data format")
				return
			}
			tunnel := Tunnel{Room1: tunnelInfo[0], Room2: tunnelInfo[1]}
			farm.Tunnels = append(farm.Tunnels, tunnel)
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	if !startRoomSet || !endRoomSet || farm.AntCount == 0 {
		fmt.Println("ERROR: invalid data format")
		return
	}

	startRoom := farm.StartRoom.Name
	endRoom := farm.EndRoom.Name
	allPaths := findAllPaths(farm.Tunnels, startRoom, endRoom)
	fmt.Println("All paths from start room to end room:")
	// allPaths dilimindeki yolları uzunluklarına göre sıralar.
	for i, path := range allPaths {
		fmt.Printf("Path %d: %v\n", i+1, path)
	}
	sort.Slice(allPaths, func(i, j int) bool {
		return len(allPaths[i]) < len(allPaths[j])
	})

	filteredPaths := filterPaths(allPaths, farm.AntCount)
	fmt.Println("Filtered Paths:")
	for i, path := range filteredPaths {
		fmt.Printf("Path %d: %v\n", i+1, path)
	}

	antMoves := simulateAntMovements(filteredPaths, farm.AntCount, startRoom, endRoom)

	fmt.Println(farm.AntCount)
	fmt.Println("##start", farm.StartRoom.Name, farm.StartRoom.X, farm.StartRoom.Y)
	fmt.Println("##end", farm.EndRoom.Name, farm.EndRoom.X, farm.EndRoom.Y)

	for _, room := range farm.Rooms {
		fmt.Printf("%s %d %d\n", room.Name, room.X, room.Y)
	}

	for _, tunnel := range farm.Tunnels {
		fmt.Printf("%s-%s\n", tunnel.Room1, tunnel.Room2)
	}

	for _, move := range antMoves {
		fmt.Println(move)
	}
	elapsed := time.Since(start)
	fmt.Println("Elapsed time:", elapsed)
}

// oda adını ve koordinatlarını içerir. Oda adı info[0] indeksindedir ve X ve Y koordinatları sırasıyla info[1] ve info[2] indekslerindedir.
func parseRoom(info []string) Room {
	x, _ := strconv.Atoi(info[1])
	y, _ := strconv.Atoi(info[2])
	return Room{Name: info[0], X: x, Y: y}
}

// Eğer oda adında boşluk karakteri varsa, bu oda adının geçersiz olduğunu belirten bir hata döndürür.
func validateRoom(room Room) error {
	if strings.Contains(room.Name, " ") {
		return fmt.Errorf("invalid room name format")
	}
	return nil
}

// tüneller aracılığıyla iki nokta arasındaki tüm geçiş yollarını bulur. dfs kullanır
func findAllPaths(tunnels []Tunnel, start, end string) [][]string {
	var paths [][]string
	visited := make(map[string]bool)
	var findPath func(current string, path []string)

	findPath = func(current string, path []string) {
		path = append(path, current)
		if current == end {
			newPath := make([]string, len(path))
			// Yolu kopyalayarak eklememizin sebebi, sonradan değişikliklerin yolda yapılan değişikliklere etkilememesini sağlamaktır.
			copy(newPath, path)
			paths = append(paths, newPath)
			return
		}
		// mevcut odanın ziyaret edildiğini işaretleriz, çünkü bu odadan geçtik.
		visited[current] = true
		// Her tünel için, eğer tünelin bir ucu mevcut odaya bağlıysa ve diğer uç daha önce ziyaret edilmediyse, bu demektir ki bu tüneli kullanarak bir sonraki odaya geçebiliriz.
		for _, tunnel := range tunnels {
			if (tunnel.Room1 == current && !visited[tunnel.Room2]) ||
				(tunnel.Room2 == current && !visited[tunnel.Room1]) {
				nextRoom := tunnel.Room1
				if tunnel.Room1 == current {
					nextRoom = tunnel.Room2
				}
				findPath(nextRoom, path)
			}
		}
		visited[current] = false
	}

	findPath(start, []string{})
	return paths
}

// Yol listesindeki yollar arasında örtüşme olmadığından emin olarak, karınca sayısına eşit veya daha az sayıda yolu seçer ve döndürür.
func filterPaths(paths [][]string, antCount int) [][]string {
	var filteredPaths [][]string
	// pathOverlap fonksiyonu, iki yolun herhangi bir ortak odayı paylaşıp paylaşmadığını belirlemek için kullanılır.
	pathsOverlap := func(path1, path2 []string) bool {
		roomsSet := make(map[string]bool)
		for _, room := range path1[1 : len(path1)-1] {
			roomsSet[room] = true
		}
		for _, room := range path2[1 : len(path2)-1] {
			if roomsSet[room] {
				return true
			}
		}
		return false
	}

	var combinations func([][]string, int, []int)
	var bestCombination []int
	maxPaths := 0

	// verilen yollar arasında uygun kombinasyonları bulmak için bir rekürsif fonksiyon olan combinations'ı tanımlıyor.
	combinations = func(paths [][]string, idx int, selected []int) {
		if len(selected) > maxPaths {
			maxPaths = len(selected)
			bestCombination = make([]int, len(selected))
			copy(bestCombination, selected)
		}
		// idx başlangıç endeksi
		for i := idx; i < len(paths); i++ {
			conflict := false
			for _, s := range selected {
				if pathsOverlap(paths[s], paths[i]) {
					conflict = true
					break
				}
			}
			// iki yol arasında ortak bir oda bulunmadığında çalışır.
			if !conflict {
				selected = append(selected, i)
				combinations(paths, i+1, selected)
				selected = selected[:len(selected)-1]
			}
		}
	}

	combinations(paths, 0, []int{})

	for _, idx := range bestCombination {
		filteredPaths = append(filteredPaths, paths[idx])
		if len(filteredPaths) == antCount {
			break
		}
	}

	return filteredPaths
}

func simulateAntMovements(paths [][]string, antCount int, start, end string) []string {
	// movements: Her turda yapılan tüm hareketleri tutan dilim.
	var movements []string
	// antPosition: Her karıncanın mevcut pozisyonunu tutan harita.
	antPosition := make(map[int]int)
	// antInEndRoom: Her karıncanın bitiş odasına ulaşıp ulaşmadığını tutan harita.
	antInEndRoom := make(map[int]bool)
	// antPaths: Her karıncanın takip edeceği yolu tutan harita.
	antPaths := make(map[int][]string)
	// activeAntCount: Halen hareket etmekte olan karıncaların sayısı.
	activeAntCount := antCount

	// Her karınca için takip edeceği yolları belirle.
	for i := 1; i <= antCount; i++ {
		if i == antCount {
			// Son karınca (antCount numaralı karınca) için yolların ilk yolunu tercih et.
			antPaths[i] = paths[0]
		} else {
			antPaths[i] = paths[(i-1)%len(paths)]
		}
		antPosition[i] = 0
		antInEndRoom[i] = false
	}

	turn := 0
	for activeAntCount > 0 {
		turn++
		// turnMovements: Bu turda yapılan hareketleri tutan dilim.
		var turnMovements []string
		// tunnelUsage: Bu turda kullanılan tünelleri tutan harita. Tünellerin aynı turda çift yönlü kullanımını engeller.
		tunnelUsage := make(map[string]bool)

		// Her karınca için mevcut odası ve bir sonraki odası belirlenir.
		// tunnel ve reverseTunnel değişkenleri, tünelin iki yönünü temsil eder.
		for i := 1; i <= antCount; i++ {
			if antInEndRoom[i] {
				continue
			}

			currentRoom := antPaths[i][antPosition[i]]
			nextRoom := antPaths[i][antPosition[i]+1]
			tunnel := fmt.Sprintf("%s-%s", currentRoom, nextRoom)
			reverseTunnel := fmt.Sprintf("%s-%s", nextRoom, currentRoom)

			if !tunnelUsage[tunnel] && !tunnelUsage[reverseTunnel] {
				turnMovements = append(turnMovements, fmt.Sprintf("L%d-%s", i, nextRoom))
				tunnelUsage[tunnel] = true
				tunnelUsage[reverseTunnel] = true
				antPosition[i]++
				if nextRoom == end {
					antInEndRoom[i] = true
					activeAntCount--
				}
			}
		}

		// Eğer hiçbir hareket yapılmazsa, bu, tüm karıncaların bitiş odasına ulaştığı veya ulaşamadığı anlamına gelir,
		// bu nedenle simülasyon sona erdirilir. Bu sayede, gereksiz döngülemelerden kaçınılmış olur ve simülasyonun daha verimli olması sağlanır.
		if len(turnMovements) > 0 {
			movements = append(movements, strings.Join(turnMovements, " "))
		} else {
			break
		}
	}
	fmt.Println("Number of turns to reach the end:", turn)

	return movements
}
