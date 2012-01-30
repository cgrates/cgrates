package timeslots

func TestSimple(*testing.T){
	
}

func BenchmarkSimple(b *testing.B) {
    for i := 0; i < b.N; i++ {
		GetCost()
    }
}